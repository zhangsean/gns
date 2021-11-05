# gns

```sh
Go network scan tool.
Version: v0.6.0

Usage: gns [Options] <IP or domain>
eg: gns -r 22-8080 -s 300 localhost

Options:
  -a    All ports, 1-65535
  -c    Show network connecting cost time
  -d    Debug, show every scan result, instead of show opening port only
  -p string
        Specify ports or port range. eg. 80,443,8080 or 80-8080 (default "21,22,23,53,80,135,139,443,445,1080,1433,1521,3306,3389,5432,6379,8080")
  -s int
        Parallel scan threads (default 200)
  -t int
        Connect timeout, ms (default 200)
```
