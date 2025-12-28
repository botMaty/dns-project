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
	store, err := storage.NewSQLiteStorage("dns_chache.db")
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	up := upstream.NewUDPUpstream("8.8.8.8:53")
	logger := &resolver.StdLogger{}
	res := resolver.New(store, up, logger)

	udp := transport.NewUDPServer(":8053", res)
	tcp := transport.NewTCPServer(":8053", res)

	doh := transport.NewDoHServer(":8054", res, "certs/cert.pem", "certs/key.pem")
	adminSrv := admin.New(store, "super-secret-token")

	mux := http.NewServeMux()
	adminSrv.Register(mux)

	go func() {
		if err := udp.ListenAndServe(); err != nil {
			log.Fatalf("UDP error: %v", err)
		}
	}()

	go func() {
		if err := tcp.ListenAndServe(); err != nil {
			log.Fatalf("TCP error: %v", err)
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
