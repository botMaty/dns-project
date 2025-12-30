package main

import (
	"dns-server/admin"
	"dns-server/resolver"
	"dns-server/storage"
	"dns-server/transport"
	"dns-server/upstream"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	udpPort := os.Getenv("UDP_PORT")
	tcpPort := os.Getenv("TCP_PORT")
	dohPort := os.Getenv("DOH_PORT")
	adminPort := os.Getenv("ADMIN_PORT")

	dohCert := os.Getenv("DOH_CERT")
	dohKey := os.Getenv("DOH_KEY")
	adminHashedPassword := os.Getenv("ADMIN_HASHED_PASSWORD")
	upstreamDNS := os.Getenv("UPSTREAM_DNS")
	databaseFile := os.Getenv("DATABASE_FILE")

	store, err := storage.NewSQLiteStorage(databaseFile)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	up := upstream.NewUDPUpstream(upstreamDNS)
	logger := &resolver.StdLogger{}
	res := resolver.New(store, up, logger)

	udp := transport.NewUDPServer(udpPort, res)
	tcp := transport.NewTCPServer(tcpPort, res)
	doh := transport.NewDoHServer(dohPort, res, dohCert, dohKey)

	adminSrv := admin.New(store, adminHashedPassword)
	mux := http.NewServeMux()
	adminSrv.Register(mux)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := store.CleanupExpired(); err != nil {
				log.Printf("cleanup error: %v", err)
			}
		}
	}()

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
		log.Fatal(http.ListenAndServe(adminPort, mux))
	}()

	select {} // برنامه زنده بماند
}
