package main

import (
	"fmt"
	"net"
)

func main() {
	addr := "localhost:21080"
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("TCP server is running at %s\n", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to establish connection.")
			continue
		}
		fmt.Printf("Accept: %s\n", conn.RemoteAddr().String())
		go func() {
			for {
				b := make([]byte, 100<<10)
				_, err := conn.Read(b)
				fmt.Printf("%s: %s", conn.RemoteAddr().String(), b)
				// conn.Write([]byte("echo: "))
				conn.Write(b)
				if err != nil {
					return
				}
			}
		}()
	}
}
