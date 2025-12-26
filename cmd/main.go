package main

import (
	"dns-server/admin"
	"dns-server/resolver"
	"dns-server/storage"
	"dns-server/transport"
	"dns-server/upstream"
	"log"
	"net/http"
)

func main() {
	store := storage.NewMemoryStorage()
	up := upstream.NewUDPUpstream("8.8.8.8:53")

	res := resolver.New(store, up)

	udp := transport.NewUDPServer(":8053", res)
	doh := transport.NewDoHServer(":8054", res, "", "")
	adminSrv := admin.New(store)

	mux := http.NewServeMux()
	adminSrv.Register(mux)

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

	go func() {
		log.Fatal(http.ListenAndServe(":8055", mux))
	}()

	select {} // برنامه زنده بماند
}
