package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/lxfontes/scanner"
)

var (
	timeout     = flag.Duration("timeout", 500*time.Millisecond, "TCP Timeout")
	endpoint    = flag.String("endpoint", "0.0.0.0:8080", "HTTP Server")
	concurrency = flag.Int("concurrency", 1024, "Concurrent connections")

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type WSPrinter struct {
	conn *websocket.Conn
	sc   *scanner.Scanner
}

type jsResult struct {
	Endpoint string `json:"endpoint"`
	PortName string `json:"port_name"`
	Open     bool   `json:"open"`
	Err      string `json:"error"`
}

func (p WSPrinter) ScanResponse(resp scanner.Response) {
	portName, _ := scanner.LookupPort(resp.Address)

	jsresp := jsResult{
		Endpoint: resp.Address.String(),
		PortName: portName,
		Open:     resp.Open,
	}

	if resp.Error != nil {
		jsresp.Err = resp.Error.Error()
	}

	err := p.conn.WriteJSON(jsresp)
	// line disconnected, don't waste our time scanning
	if err != nil {
		log.Println("Stopping scan (lost socket)")
		p.sc.Stop()
	}
}

func wsConnection(w http.ResponseWriter, r *http.Request) {
	rawIP := r.FormValue("ip")
	rawStart := r.FormValue("start")
	rawEnd := r.FormValue("end")

	if rawIP == "" {
		return
	}

	ip := net.ParseIP(rawIP)

	start, err := strconv.Atoi(rawStart)
	if err != nil {
		log.Println(err)
		return
	}

	end, err := strconv.Atoi(rawEnd)
	if err != nil {
		log.Println(err)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Starting scan on",
		ip.String(),
		start,
		"-",
		end,
		"for client",
		r.RemoteAddr,
	)

	go streamScan(conn, ip, start, end)
}

func streamScan(conn *websocket.Conn, ip net.IP, start int, end int) {
	wsPrinter := WSPrinter{
		conn: conn,
	}

	cfg := scanner.ScannerConfig{
		Timeout:     *timeout,
		Concurrency: *concurrency,
		Reporter:    &wsPrinter,
	}

	sc := scanner.NewScanner(cfg)
	wsPrinter.sc = sc

	sc.ScanIPPortRange(ip.String(), start, end)
	conn.Close()
}

func main() {
	flag.Parse()

	r := mux.NewRouter()
	r.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/scan", wsConnection)
	http.Handle("/", r)
	http.ListenAndServe(*endpoint, nil)
}
