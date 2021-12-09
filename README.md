# gns

A pretty fast network scan tool written with golang, it scans all opening ports on the target network.

[![Codecov](https://codecov.io/gh/zhangsean/gns/branch/master/graph/badge.svg)](https://codecov.io/gh/zhangsean/gns)
[![Go Report Card](https://goreportcard.com/badge/github.com/zhangsean/gns)](https://goreportcard.com/report/github.com/zhangsean/gns)

## Quick view

[![asciicast](https://asciinema.org/a/448361.svg)](https://asciinema.org/a/448361)

## Install

#### With go
```sh
go get -u github.com/zhangsean/gns
```

#### Manual install
Download build from [releases](https://github.com/zhangsean/gns/releases/latest)

## Usage

```sh
Go network scan tool.
Version: v0.8.4

Usage: gns [Options] <IP range or domain>
eg: gns -p 22-8080 -s 300 10.0.1.1-100
    gns -p 80,443 -t 500 10.0.1.0/24
    gns www.google.com

Options:
  -a    All ports, 1-65535
  -c    Show network connecting cost time
  -d    Debug, show every scan result, instead of showing open port only
  -h    Show help
  -p string
        Specify ports or port range. eg. 80,443,8080 or 80-8080 (default "21,22,23,53,80,135,139,443,445,1080,1433,1521,3306,3389,5432,6379,8080")
  -s int
        Parallel scan threads (default 200)
  -t int
        Connect timeout, ms (default 1000)
  -v    Show version
```
