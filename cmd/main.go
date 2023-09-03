package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/atori74/bridge"
)

func main() {
	var (
		mode     = flag.String("mode", "server", "running mode. One of (server, client).")
		laddr    = flag.String("local", "localhost:28080", "Local address that listen on.")
		raddr    = flag.String("remote", "localhost:28081", "Remote address that connect to")
		protocol = flag.String("protocol", "wss", "Protocol to communicate with remote server.")
	)
	flag.Parse()

	switch *mode {
	case "client":
		t := bridge.TCPListener{
			Address:        *laddr,
			RemoteAddress:  *raddr,
			RemoteProtocol: *protocol,
		}
		t.Listen()
	case "server":
		h := bridge.HTTPListener{
			Address:       *laddr,
			RemoteAddress: *raddr,
		}
		h.Listen()
	default:
		fmt.Println("Specify mode of client or server as option.")
		os.Exit(1)
	}
}
