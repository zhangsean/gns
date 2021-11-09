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

const ver string = "v0.8.1"

var ports string
var parallels int
var all bool
var showCostTime bool
var debug bool
var warning bool
var ms int64
var mutex sync.Mutex
var showVer bool
var showHelp bool

type TCPAddrStatus struct {
	Addr   net.TCPAddr
	Status string
}

var resAddrs []TCPAddrStatus

func AppendStatus(tcpStatus TCPAddrStatus) {
	mutex.Lock()
	resAddrs = append(resAddrs, tcpStatus)
	mutex.Unlock()
}

func IPStringToInt(ipString string) int {
	ipSegs := strings.Split(ipString, ".")
	var ipInt int = 0
	var pos uint = 24
	for _, ipSeg := range ipSegs {
		tempInt, _ := strconv.Atoi(ipSeg)
		tempInt = tempInt << pos
		ipInt = ipInt | tempInt
		pos -= 8
	}
	return ipInt
}

func IPIntToString(ipInt int) string {
	ipSecs := make([]string, 4)
	for i := 0; i < 4; i++ {
		ipSecs[3-i] = strconv.Itoa(ipInt & 0xFF)
		ipInt >>= 8
	}
	return strings.Join(ipSecs, ".")
}

func AppendPort(ports []int, port int) []int {
	if port > 0 && port < 65536 {
		ports = append(ports, port)
	}
	return ports
}

func init() {
	flag.BoolVar(&all, "a", false, "All ports, 1-65535")
	flag.BoolVar(&showCostTime, "c", false, "Show network connecting cost time")
	flag.BoolVar(&debug, "d", false, "Debug, show every scan result, instead of show opening port only")
	flag.StringVar(&ports, "p", "21,22,23,53,80,135,139,443,445,1080,1433,1521,3306,3389,5432,6379,8080", "Specify ports or port range. eg. 80,443,8080 or 80-8080")
	flag.IntVar(&parallels, "s", 200, "Parallel scan threads")
	flag.Int64Var(&ms, "t", 200, "Connect timeout, ms")
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showVer, "v", false, "Show version")
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Go network scan tool.\nVersion: "+ver+"\n")
		if !showVer {
			fmt.Fprintf(os.Stdout, "\nUsage: gns [Options] <IP range or domain>\neg: gns -p 22-8080 -s 300 10.0.1.1-100\n\nOptions:\n")
			flag.PrintDefaults()
		}
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
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("[%v]:%v", tcpAddr.IP, tcpAddr.Port), time.Duration(ms*int64(time.Millisecond)))
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
	if showHelp || showVer {
		flag.Usage()
		return
	}
	var aimIPs []net.IP
	host := flag.Arg(0)
	if len(host) == 0 {
		flag.Usage()
		return
	}
	if ip := net.ParseIP(host); ip != nil {
		aimIPs = append(aimIPs, ip)
	} else if ip, ipNet, _ := net.ParseCIDR(host); ipNet != nil {
		fmt.Printf("IP CIDR: %v, %v\n", ip, ipNet)
		for i := IPStringToInt(ip.String()); i > 0; i++ {
			tmpIP := net.ParseIP(IPIntToString(i))
			if ipNet.IP.Equal(tmpIP) {
				continue
			} else if ipNet.Contains(tmpIP) && !tmpIP.IsLoopback() && !tmpIP.IsMulticast() && !tmpIP.IsUnspecified() {
				aimIPs = append(aimIPs, tmpIP)
			} else {
				break
			}
		}
	} else if isIPRange, _ := regexp.Match(`^(\d{1,3}\.){3}\d{1,3}-\d{1,3}$`, []byte(host)); isIPRange {
		ipSpecs := strings.Split(host, ".")
		rangeSpecs := strings.Split(ipSpecs[3], "-")
		startNum, _ := strconv.Atoi(rangeSpecs[0])
		endNum, _ := strconv.Atoi(rangeSpecs[1])
		if startNum > endNum {
			tmp := endNum
			endNum = startNum
			startNum = tmp
		}
		ipSpecs = ipSpecs[:len(ipSpecs)-1]
		for i := startNum; i <= endNum; i++ {
			tmpIPSpecs := append(ipSpecs, strconv.Itoa(i))
			if tmpIp := net.ParseIP(strings.Join(tmpIPSpecs, ".")); tmpIp != nil {
				aimIPs = append(aimIPs, tmpIp)
			}
		}
	} else {
		ips, err := net.LookupIP(host)
		if err == nil && len(ips) > 0 {
			fmt.Printf("Resolving %v => %v\n", host, ips)
			aimIPs = append(aimIPs, ips[0])
		}
	}
	if len(aimIPs) == 0 {
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

	var aimPorts []int
	if isRange {
		if startPort > endPort {
			tmpPort := endPort
			endPort = startPort
			startPort = tmpPort
		}
		for i := startPort; i <= endPort; i++ {
			aimPorts = AppendPort(aimPorts, i)
		}
	} else {
		for _, p := range strings.Split(ports, ",") {
			port, _ := strconv.Atoi(p)
			aimPorts = AppendPort(aimPorts, port)
		}
	}
	if len(aimPorts) == 0 {
		fmt.Fprintln(os.Stderr, "No valid network port to scan, port must between 1 and 65535.")
		return
	}

	msg := "Scaning port"
	if strings.Contains(ports, ",") || strings.Contains(ports, "-") {
		msg += "s "
	} else {
		msg += " "
	}
	msg += ports + " on"
	fmt.Println(msg, aimIPs)

	scanCount := len(aimIPs) * len(aimPorts)
	wg := sync.WaitGroup{}
	wg.Add(scanCount)
	bar := pb.StartNew(scanCount)
	bar.ShowTimeLeft = true
	parallelChan := make(chan int, parallels)
	for _, ip := range aimIPs {
		for _, port := range aimPorts {
			parallelChan <- 1
			go checkPort(ip, port, &wg, parallelChan, bar)
		}
	}
	wg.Wait()
	bar.Finish()

	fmt.Println("----Scan Result----")
	if !debug && len(resAddrs) == 0 {
		fmt.Println("No opening port")
	} else {
		sort.SliceStable(resAddrs, func(i, j int) bool {
			return IPStringToInt(resAddrs[i].Addr.IP.String())*65536+resAddrs[i].Addr.Port < IPStringToInt(resAddrs[j].Addr.IP.String())*65536+resAddrs[j].Addr.Port
		})
		for _, t := range resAddrs {
			fmt.Println(t.Addr.String() + " : " + t.Status)
		}
	}
	if warning {
		fmt.Fprintln(os.Stderr, "Warning: too many open sockets, please slow down.")
	}
}
