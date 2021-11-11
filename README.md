# gns

Pretty fast network scan tool for all platform written in Go, scanning all opening port in your aim subnet.

[![Codecov](https://codecov.io/gh/zhangsean/gns/branch/master/graph/badge.svg)](https://codecov.io/gh/zhangsean/gns)
[![Go Report Card](https://goreportcard.com/badge/github.com/zhangsean/gns)](https://goreportcard.com/report/github.com/zhangsean/gns)
[![Maintainability](https://api.codeclimate.com/v1/badges/30d793274607f599e658/maintainability)](https://codeclimate.com/github/zhangsean/gns/maintainability)
[![CodeScene Code Health](https://codescene.io/projects/13129/status-badges/code-health)](https://codescene.io/projects/13129)

[![asciicast](https://asciinema.org/a/448361.svg)](https://asciinema.org/a/448361)

## Install

```sh
go get -u github.com/zhangsean/gns
```

## Usage

```sh
Go network scan tool.
Version: v0.8.2

Usage: gns [Options] <IP range or domain>
eg: gns -p 22-8080 -s 300 10.0.1.1-100
eg: gns -p 80,443 -t 500 10.0.1.0/24

Options:
  -a    All ports, 1-65535
  -c    Show network connecting cost time
  -d    Debug, show every scan result, instead of show opening port only
  -h    Show help
  -p string
        Specify ports or port range. eg. 80,443,8080 or 80-8080 (default "21,22,23,53,80,135,139,443,445,1080,1433,1521,3306,3389,5432,6379,8080")
  -s int
        Parallel scan threads (default 200)
  -t int
        Connect timeout, ms (default 200)
  -v    Show version
```
