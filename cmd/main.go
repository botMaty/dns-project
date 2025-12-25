package main

import (
	"dns-server/resolver"
	"dns-server/storage"
	"dns-server/transport"
	"dns-server/upstream"
	"log"
)

func main() {
	// 1️⃣ ساخت Storage
	memStore := storage.NewMemoryStorage()

	// 2️⃣ ساخت Upstream (مثلاً Google DNS)
	up := upstream.NewUDPUpstream("8.8.8.8:53")

	// 3️⃣ ساخت Resolver
	res := resolver.New(memStore, up)

	// 4️⃣ ساخت UDP Server روی پورت 8053
	udp := transport.NewUDPServer(":8053", res)

	// 5️⃣ شروع به کار سرور
	log.Println("Starting UDP DNS server on :8053")
	if err := udp.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
