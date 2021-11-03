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

const ver string = "v0.4.0"

var ports string
var portRange string
var parallels int
var all bool
var debug bool
var warning bool
var ms int64
var scanCount int
var mutex sync.Mutex

type TCPAddrStatus struct {
	Addr   net.TCPAddr
	Status string
}

var tcpAddrs []TCPAddrStatus

func Append(tcpStatus TCPAddrStatus) {
	mutex.Lock()
	tcpAddrs = append(tcpAddrs, tcpStatus)
	mutex.Unlock()
}

func init() {
	flag.BoolVar(&all, "a", false, "All ports, 1-65535")
	flag.BoolVar(&debug, "d", false, "Debug, show every scan result, instead of show opening port only")
	flag.StringVar(&ports, "p", "21,22,23,53,80,135,139,443,445,1080,1433,1521,2222,3000,3306,3389,5432,6379,8080,8888,50050,55553", "Specify ports")
	flag.StringVar(&portRange, "r", "", "Range ports, <from>-<to>. eg. 80-8080")
	flag.IntVar(&parallels, "s", 200, "Parallel scan threads")
	flag.Int64Var(&ms, "t", 200, "Connect timeout, ms")
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Go network scan tool.\nVersion: "+ver+"\n\nUsage: gns [Options] <IP>\neg: gns -r 22-8080 -s 300 127.0.0.1\n\nOptions:\n")
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
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", tcpAddr.IP, tcpAddr.Port), time.Duration(ms*int64(time.Millisecond)))
	if err == nil {
		Append(TCPAddrStatus{tcpAddr, "opening"})
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
		if debug {
			Append(TCPAddrStatus{tcpAddr, errMsg})
		}
	}
	bar.Increment()
	<-parallelChan
}

func main() {
	args := flag.Args()
	ip := net.ParseIP(flag.Arg(0))
	if len(args) != 1 || ip == nil || strings.Contains(ports, "-") {
		flag.Usage()
	} else {
		if all {
			portRange = "1-65535"
		}
		msg := "Scaning port"
		if portRange != "" {
			msg += "s " + portRange
		} else {
			if strings.Contains(ports, ",") {
				msg += "s "
			} else {
				msg += " "
			}
			msg += ports
		}
		msg += " on " + ip.String()
		fmt.Println(msg)
		wg := sync.WaitGroup{}
		if portRange != "" {
			matched, _ := regexp.Match(`^\d+-\d+$`, []byte(portRange))
			if !matched {
				flag.Usage()
			} else {
				portSecs := strings.Split(portRange, "-")
				startPort, err1 := strconv.Atoi(portSecs[0])
				endPort, err2 := strconv.Atoi(portSecs[1])
				if err1 != nil || err2 != nil || startPort < 1 || endPort < 2 || endPort <= startPort || parallels < 1 {
					flag.Usage()
				} else {
					scanCount = endPort - startPort + 1
					wg.Add(scanCount)
					bar := pb.StartNew(scanCount)
					bar.ShowTimeLeft = true
					parallelChan := make(chan int, parallels)
					for i := startPort; i <= endPort; i++ {
						parallelChan <- 1
						go checkPort(ip, i, &wg, parallelChan, bar)
					}
					wg.Wait()
					bar.Finish()
				}
			}
		} else {
			parallelChan := make(chan int, parallels)
			arr := strings.Split(ports, ",")
			scanCount = len(arr)
			wg.Add(scanCount)
			bar := pb.StartNew(scanCount)
			bar.ShowTimeLeft = true
			for i := 0; i < scanCount; i++ {
				p, err := strconv.Atoi(arr[i])
				if err == nil {
					parallelChan <- 1
					go checkPort(ip, p, &wg, parallelChan, bar)
				}
			}
			wg.Wait()
			bar.Finish()
		}
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
			fmt.Fprintf(os.Stderr, "Warning: too many open sockets, please slow down.")
		}
	}
}
