package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cheggaaa/pb"
)

const ver string = "v0.6.0"

var ports string
var parallels int
var all bool
var showCostTime bool
var debug bool
var warning bool
var ms int64
var mutex sync.Mutex

type TCPAddrStatus struct {
	Addr   net.TCPAddr
	Status string
}

var tcpAddrs []TCPAddrStatus

func AppendStatus(tcpStatus TCPAddrStatus) {
	mutex.Lock()
	tcpAddrs = append(tcpAddrs, tcpStatus)
	mutex.Unlock()
}

var portList []int

func AppendPort(port int) {
	if port > 0 && port < 65536 {
		portList = append(portList, port)
	}
}

func init() {
	flag.BoolVar(&all, "a", false, "All ports, 1-65535")
	flag.BoolVar(&showCostTime, "c", false, "Show network connecting cost time")
	flag.BoolVar(&debug, "d", false, "Debug, show every scan result, instead of show opening port only")
	flag.StringVar(&ports, "p", "21,22,23,53,80,135,139,443,445,1080,1433,1521,3306,3389,5432,6379,8080", "Specify ports or port range. eg. 80,443,8080 or 80-8080")
	flag.IntVar(&parallels, "s", 200, "Parallel scan threads")
	flag.Int64Var(&ms, "t", 200, "Connect timeout, ms")
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Go network scan tool.\nVersion: "+ver+"\n\nUsage: gns [Options] <IP or domain>\neg: gns -r 22-8080 -s 300 localhost\n\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
}

func checkPort(ip net.IP, port int, wg *sync.WaitGroup, parallelChan chan int, bar *pb.ProgressBar) {
	defer wg.Done()
	tcpAddr := net.TCPAddr{
		IP:   ip,
		Port: port,
	}
	timeStart := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", tcpAddr.IP, tcpAddr.Port), time.Duration(ms*int64(time.Millisecond)))
	timeCost := time.Since(timeStart).String()
	if err == nil {
		msg := "opening"
		if showCostTime {
			msg += " " + timeCost
		}
		AppendStatus(TCPAddrStatus{tcpAddr, msg})
		conn.Close()
	} else {
		errMsg := err.Error()
		if strings.Contains(errMsg, "connection refused") {
			errMsg = "refused"
		} else if strings.Contains(errMsg, "connect: operation timed out") || strings.Contains(errMsg, "i/o timeout") {
			errMsg = "timeout"
		} else if strings.Contains(errMsg, "socket: too many open files") {
			warning = true
			errMsg = "retrying"
			wg.Add(1)
			parallelChan <- 1
			checkPort(ip, port, wg, parallelChan, bar)
		}
		if showCostTime {
			errMsg += " " + timeCost
		}
		if debug {
			AppendStatus(TCPAddrStatus{tcpAddr, errMsg})
		}
	}
	bar.Increment()
	<-parallelChan
}

func main() {
	host := flag.Arg(0)
	ip := net.ParseIP(host)
	if ip == nil && len(host) > 0 {
		addrs, err := net.LookupHost(host)
		if err == nil {
			fmt.Printf("%v => %v\n", host, addrs)
			ip = net.ParseIP(addrs[0])
		}
	}
	if len(flag.Args()) != 1 || ip == nil {
		fmt.Fprintln(os.Stderr, "Invalid IP, hostname or domain")
		return
	}

	if all {
		ports = "1-65535"
	}
	isList, _ := regexp.Match(`^\d[,\d]*$`, []byte(ports))
	isRange, _ := regexp.Match(`^\d+-\d+$`, []byte(ports))
	var startPort, endPort int
	if isRange {
		portSecs := strings.Split(ports, "-")
		startPort, _ = strconv.Atoi(portSecs[0])
		endPort, _ = strconv.Atoi(portSecs[1])
	}
	if !isList && !isRange {
		fmt.Fprintln(os.Stderr, "Invalid ports")
		return
	}

	if isRange {
		if startPort < endPort {
			for i := startPort; i <= endPort; i++ {
				AppendPort(i)
			}
		} else {
			for i := endPort; i <= startPort; i++ {
				AppendPort(i)
			}
		}
	} else {
		for _, p := range strings.Split(ports, ",") {
			port, _ := strconv.Atoi(p)
			AppendPort(port)
		}
	}
	if len(portList) == 0 {
		fmt.Fprintln(os.Stderr, "No valid network port to scan, port must between 1 and 65535.")
		return
	}

	msg := "Scaning port"
	if strings.Contains(ports, ",") || strings.Contains(ports, "-") {
		msg += "s "
	} else {
		msg += " "
	}
	msg += ports + " on " + ip.String()
	fmt.Println(msg)

	wg := sync.WaitGroup{}
	wg.Add(len(portList))
	bar := pb.StartNew(len(portList))
	bar.ShowTimeLeft = true
	parallelChan := make(chan int, parallels)
	for _, port := range portList {
		parallelChan <- 1
		go checkPort(ip, port, &wg, parallelChan, bar)
	}
	wg.Wait()
	bar.Finish()

	fmt.Println("----Scan Result----")
	if !debug && len(tcpAddrs) == 0 {
		fmt.Println("No opening port")
	} else {
		sort.Slice(tcpAddrs, func(i, j int) bool {
			return tcpAddrs[i].Addr.Port < tcpAddrs[j].Addr.Port
		})
		for _, t := range tcpAddrs {
			fmt.Println("Port " + strconv.Itoa(t.Addr.Port) + " is " + t.Status)
		}
	}
	if warning {
		fmt.Fprintln(os.Stderr, "Warning: too many open sockets, please slow down.")
	}
}
