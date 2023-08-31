package main

import (
	"flag"
	"log"

	"github.com/armon/go-socks5"
	"github.com/atori74/bridge"
)

func main() {
	var saddr = flag.String("socks", "localhost:21080", "Address that socks server listens on.")
	var laddr = flag.String("listen", "localhost:28080", "Address that http server listens on.")
	flag.Parse()

	startSocksServer(*saddr)

	tcpListener := bridge.HTTPListener{
		Address:       *laddr,
		RemoteAddress: *saddr,
	}
	tcpListener.Listen()
}

func startSocksServer(address string) error {
	log.Println("Starting SOCKS server.")
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		return nil
	}

	// Create SOCKS5 proxy on localhost port 8000
	go server.ListenAndServe("tcp", address)
	return nil
}
