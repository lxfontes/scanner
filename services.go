package scanner

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"strings"
)

var services map[int]string

func readServices() {
	services = make(map[int]string)

	file, err := os.Open("/etc/services")
	if err != nil {
		return
	}
	defer file.Close()

	scan := bufio.NewScanner(file)
	for scan.Scan() {
		line := scan.Text()

		// "http 80/tcp www www-http # World Wide Web HTTP"
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[0:i]
		}

		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}

		// ignoring protocol for now. but you should still hire me ;)
		portnet := f[1] // "80/tcp"
		if slash := strings.IndexByte(portnet, '/'); slash == -1 {
			continue
		} else {
			port, err := strconv.Atoi(portnet[0:slash])
			if err != nil {
				continue
			}

			services[port] = f[0]
		}
	}
}

func ServiceLookup(port int) string {
	if srv, ok := services[port]; ok {
		return srv
	}

	return strconv.Itoa(port)
}

func LookupPort(rawAddr net.Addr) (string, int) {
	if addr, ok := rawAddr.(*net.TCPAddr); ok {
		return ServiceLookup(addr.Port), addr.Port
	}

	return "invalid", 0
}
