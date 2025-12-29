package main

import (
	"bytes"
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
	useHTTPS := flag.Bool("https", false, "use HTTPS")
	method := flag.String("method", "get", "HTTP method: get or post")
	flag.Parse()

	if *name == "" {
		fmt.Println("domain name required")
		os.Exit(1)
	}

	packet, err := buildQuery(*name + ".")
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	if *useHTTPS {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // accept self-signed
			},
		}
		client.Transport = tr
	}

	var resp *http.Response
	scheme := "http"
	if *useHTTPS {
		scheme = "https"
	}

	switch *method {
	case "get":
		dnsParam := base64.RawURLEncoding.EncodeToString(packet)
		fullURL := scheme + "://" + *url + "?dns=" + dnsParam
		resp, err = client.Get(fullURL)

	case "post":
		fullURL := scheme + "://" + *url
		req, errReq := http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(packet))
		if errReq != nil {
			panic(errReq)
		}
		req.Header.Set("Content-Type", "application/dns-message")
		resp, err = client.Do(req)

	default:
		fmt.Println("invalid method: use get or post")
		os.Exit(1)
	}

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
		// resource, ok := a.Body.(*dnsmessage.AResource) // برای دسترسی به مقدار
	}
}
