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
	"time"
)

var port string
var portRange string
var parallelCounts int
var all bool

func init() {
	flag.BoolVar(&all, "a", false, "all ports, 1-65535")
	flag.StringVar(&port, "p", "21,22,23,53,80,135,139,443,445,1080,1433,1521,2222,3000,3306,3389,8080,8888,50050,55553", "ports")
	flag.StringVar(&portRange, "r", "", "range ports, <from>-<to>. eg. 80-8080")
	flag.IntVar(&parallelCounts, "s", 200, "parallel scan threads")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: [Options] <IP>\n\neg: ./portscan -r 1-65535 -s 10000 127.0.0.1\n\nOptions:\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
}

func printPort(port int, state string) {
	fmt.Println("port " + strconv.Itoa(port) + " is " + state)
}

func checkPort(ip net.IP, port int, wg *sync.WaitGroup, parallelChan *chan int) {
	defer wg.Done()
	tcpAddr := net.TCPAddr{
		IP:   ip,
		Port: port,
	}
	conn, err := net.DialTCP("tcp", nil, &tcpAddr)
	if err == nil {
		printPort(port, "opening")
		conn.Close()
	} else if strings.Contains(err.Error(), "connection refused") {
		// printPort(port, "refused")
	} else if strings.Contains(err.Error(), "socket: too many open files") {
		printPort(port, "retrying")
		time.Sleep(time.Second)
		checkPort(ip, port, wg, parallelChan)
	} else {
		printPort(port, err.Error())
	}
	<-*parallelChan
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
						go checkPort(ip, i, &wg, &parallelChan)
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
					go checkPort(ip, p, &wg, &parallelChan)
				}
			}
			wg.Wait()
		}
	}
}
