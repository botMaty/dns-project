package main

import (
	"dns-server/resolver"
	"dns-server/storage"
	"dns-server/transport"
	"dns-server/upstream"
	"log"
)

func main() {
	store := storage.NewMemoryStorage()
	up := upstream.NewUDPUpstream("8.8.8.8:53")

	res := resolver.NewResolver(store, up)

	udp := transport.NewUDPServer(":8053", res)
	doh := transport.NewDoHServer(":8054", res, "", "")

	go func() {
		if err := udp.ListenAndServe(); err != nil {
			log.Fatalf("UDP error: %v", err)
		}
	}()

	go func() {
		if err := doh.ListenAndServe(); err != nil {
			log.Fatalf("DoH error: %v", err)
		}
	}()

	select {} // برنامه زنده بماند
}
