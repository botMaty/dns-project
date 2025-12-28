package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"crypto/tls"

	"golang.org/x/net/dns/dnsmessage"
)

func buildQuery(name string) ([]byte, error) {
	msg := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID: 1,
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
	url := flag.String("url", "127.0.0.1:8054/dns-query", "DoH endpoint host:port (without scheme)")
	name := flag.String("name", "", "domain name")
	useHTTPS := flag.Bool("https", false, "use HTTPS instead of HTTP")
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
	scheme := "http"
	if *useHTTPS {
		scheme = "https"
	}
	fullURL := scheme + "://" + *url + "?dns=" + dnsParam

	client := &http.Client{}
	if *useHTTPS {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // accept self-signed
			},
		}
		client.Transport = tr
	}

	resp, err := client.Get(fullURL)
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

	if h.RCode != dnsmessage.RCodeSuccess {
		return
	}

	p.SkipAllQuestions()

	answers, _ := p.AllAnswers()
	for _, a := range answers {
		fmt.Println(a.Header.Name, a.Header.Type, a.Header.TTL)
	}
	// get value: resource, ok := a.Body.(*dnsmessage.{STR TYPE}Resource)
}
