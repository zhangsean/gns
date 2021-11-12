package main

import (
	"net"
	"reflect"
	"sync"
	"testing"

	"github.com/cheggaaa/pb"
)

func TestAppendStatus(t *testing.T) {
	bc := len(resAddrs)
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			AppendStatus(TCPAddrStatus{})
			wg.Done()
		}(wg)
	}
	wg.Wait()
	if ac := len(resAddrs) - bc; ac != 10 {
		t.Errorf("Append 10 TCPAddrStatus, but got %v", ac)
	}
}

func TestIPStringToInt(t *testing.T) {
	type args struct {
		ipString string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"10.1.1.1", args{"10.1.1.1"}, 167837953},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IPStringToInt(tt.args.ipString); got != tt.want {
				t.Errorf("IPStringToInt(%v) = %v, want %v", tt.args.ipString, got, tt.want)
			}
		})
	}
}

func TestIPIntToString(t *testing.T) {
	type args struct {
		ipInt int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"10.1.1.1", args{167837953}, "10.1.1.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IPIntToString(tt.args.ipInt); got != tt.want {
				t.Errorf("IPIntToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppendPort(t *testing.T) {
	type args struct {
		ports []int
		port  int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"0", args{[]int{}, 0}, []int{}},
		{"65536", args{[]int{}, 65536}, []int{}},
		{"80", args{[]int{}, 80}, []int{80}},
		{"65535", args{[]int{}, 65535}, []int{65535}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AppendPort(tt.args.ports, tt.args.port); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppendPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPort(t *testing.T) {
	type args struct {
		ip           net.IP
		port         int
		wg           *sync.WaitGroup
		parallelChan chan int
		bar          *pb.ProgressBar
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	pc := make(chan int, 1)
	pc <- 1
	pb := pb.New(1)
	tests := []struct {
		name string
		args args
	}{
		{"first", args{net.ParseIP("127.0.0.1"), 1, wg, pc, pb}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cnt := len(resAddrs)
			CheckPort(tt.args.ip, tt.args.port, tt.args.wg, tt.args.parallelChan, tt.args.bar)
			if got := len(resAddrs) - cnt; got != 0 {
				t.Errorf("checkPort() = %v, want %v", got, 0)
			}
		})
	}
}
