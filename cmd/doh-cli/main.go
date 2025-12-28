package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/net/dns/dnsmessage"
)

func buildQuery(name string) ([]byte, error) {
	msg := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:       1,
			Response: false,
		},
		Questions: []dnsmessage.Question{
			{
				Name:  dnsmessage.MustNewName(name),
				Type:  dnsmessage.TypeA,
				Class: dnsmessage.ClassINET,
			},
		},
	}
	return msg.Pack()
}

func main() {
	url := flag.String("url", "http://127.0.0.1:8054/query-dns/", "DoH endpoint")
	name := flag.String("name", "", "domain name")
	flag.Parse()

	if *name == "" {
		fmt.Println("domain name required")
		os.Exit(1)
	}

	packet, err := buildQuery(*name + ".")
	if err != nil {
		panic(err)
	}

	dnsParam := base64.RawURLEncoding.EncodeToString(packet)
	fullURL := *url + "?dns=" + dnsParam

	resp, err := http.Get(fullURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var p dnsmessage.Parser
	h, err := p.Start(body)
	if err != nil {
		panic(err)
	}

	fmt.Println("RCode:", h.RCode)
	answers, _ := p.AllAnswers()
	for _, a := range answers {
		fmt.Println(a.Header.Name, a.Header.Type)
	}

}
