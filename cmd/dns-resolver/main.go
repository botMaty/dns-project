package main

import (
	"fmt"
	"local-dns/pkg/dns"
	"net"
)

func main() {
	fmt.Printf("Starting DNS Server...\n")
	packetConnection, err := net.ListenPacket("udp", ":8053")
	if err != nil {
		panic(err)
	}
	defer packetConnection.Close()
	for {
		buf := make([]byte, 512)
		n, addr, err := packetConnection.ReadFrom(buf)
		if err != nil {
			fmt.Printf("Read error from %s: %s\n", addr.String(), err)
			continue
		}
		go dns.HandlePacket(packetConnection, addr, buf[:n])
	}
}
