# Simple TCP Scanner

- library approach
- ipv4
- ipv6
- cidr
- concurrency
- timeout

## cli

```
go get github.com/lxfontes/cmd/scanner
$ scanner [options] [cidr] [cidr] [cidr]
Example
   ./scanner 127.0.0.1/32 192.168.88.0/24 2001:DB8::/48

Options
  -concurrency int
      Concurrent connections (default 10)
  -end int
      Port Range End (default 65535)
  -start int
      Port Range Start (default 1)
  -timeout duration
      TCP Timeout (default 1s)
  -verbose
      Verbose Output
```

## web

Check [here](cmd/webscanner).

Live version at
[http://scanner.adorablehacker.com](http://scanner.adorablehacker.com).
