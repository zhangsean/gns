package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const ver string = "v0.2.1"

var port string
var portRange string
var parallelCounts int
var all bool
var debug bool
var warning bool

func init() {
	flag.BoolVar(&all, "a", false, "All ports, 1-65535")
	flag.BoolVar(&debug, "d", false, "Debug, show every scan result, instead of show opening port only")
	flag.StringVar(&port, "p", "21,22,23,53,80,135,139,443,445,1080,1433,1521,2222,3000,3306,3389,5432,6379,8080,8888,50050,55553", "Specify ports")
	flag.StringVar(&portRange, "r", "", "Range ports, <from>-<to>. eg. 80-8080")
	flag.IntVar(&parallelCounts, "s", 200, "Parallel scan threads")
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Go network scan tool.\nVersion: "+ver+"\n\nUsage: gns [Options] <IP>\neg: gns -r 22-8080 -s 300 127.0.0.1\n\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
}

func printPort(port int, state string) {
	fmt.Println("port " + strconv.Itoa(port) + " is " + state)
}

func checkPort(ip net.IP, port int, wg *sync.WaitGroup, parallelChan chan int) {
	defer wg.Done()
	tcpAddr := net.TCPAddr{
		IP:   ip,
		Port: port,
	}
	conn, err := net.DialTCP("tcp", nil, &tcpAddr)
	if err == nil {
		printPort(port, "opening")
		conn.Close()
	} else {
		errMsg := err.Error()
		if strings.Contains(errMsg, "connection refused") {
			errMsg = "refused"
		} else if strings.Contains(errMsg, "connect: operation timed out") {
			errMsg = "timeout"
		} else if strings.Contains(errMsg, "socket: too many open files") {
			warning = true
			errMsg = "retrying"
			wg.Add(1)
			parallelChan <- 1
			checkPort(ip, port, wg, parallelChan)
		}
		if debug {
			printPort(port, errMsg)
		}
	}
	<-parallelChan
}

func main() {
	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
	} else {
		ip := net.ParseIP(flag.Arg(0))
		if all {
			portRange = "1-65535"
		}
		wg := sync.WaitGroup{}
		if portRange != "" {
			matched, _ := regexp.Match(`^\d+-\d+$`, []byte(portRange))
			if !matched {
				flag.Usage()
			} else {
				portSecs := strings.Split(portRange, "-")
				startPort, err1 := strconv.Atoi(portSecs[0])
				endPort, err2 := strconv.Atoi(portSecs[1])
				if err1 != nil || err2 != nil || startPort < 1 || endPort < 2 || endPort <= startPort || parallelCounts < 1 {
					flag.Usage()
				} else {
					wg.Add(endPort - startPort + 1)
					parallelChan := make(chan int, parallelCounts)
					for i := startPort; i <= endPort; i++ {
						parallelChan <- 1
						go checkPort(ip, i, &wg, parallelChan)
					}
					wg.Wait()
				}
			}
		} else {
			parallelChan := make(chan int, parallelCounts)
			arr := strings.Split(port, ",")
			wg.Add(len(arr))
			for i := 0; i < len(arr); i++ {
				p, err := strconv.Atoi(arr[i])
				if err == nil {
					parallelChan <- 1
					go checkPort(ip, p, &wg, parallelChan)
				}
			}
			wg.Wait()
		}
		if warning {
			fmt.Fprintf(os.Stderr, "Warning: too many open sockets, please slow down.")
		}
	}
}
