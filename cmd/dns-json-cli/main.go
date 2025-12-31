package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	url := flag.String("url", "127.0.0.1:8054/dns-query/json", "DNS JSON endpoint host:port (without scheme)")
	name := flag.String("name", "", "domain name")
	type_ := flag.String("type", "A", "record type (A, AAAA, CNAME, etc.)")
	method := flag.String("method", "get", "HTTP method: get or post")
	useHTTPS := flag.Bool("https", false, "use HTTPS")
	flag.Parse()

	if *name == "" {
		fmt.Println("domain name required")
		os.Exit(1)
	}

	client := &http.Client{}
	if *useHTTPS {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // accept self-signed
		}
	}

	scheme := "http"
	if *useHTTPS {
		scheme = "https"
	}

	var resp *http.Response
	var err error

	switch *method {
	case "get":
		fullURL := fmt.Sprintf("%s://%s?name=%s&type=%s", scheme, *url, *name, *type_)
		resp, err = client.Get(fullURL)

	case "post":
		bodyJSON, _ := json.Marshal(map[string]string{
			"name": *name,
			"type": *type_,
		})
		req, _ := http.NewRequest(http.MethodPost, scheme+"://"+*url, bytes.NewReader(bodyJSON))
		req.Header.Set("Content-Type", "application/dns-json")
		resp, err = client.Do(req)

	default:
		fmt.Println("invalid method: use get or post")
		os.Exit(1)
	}

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		panic(err)
	}

	pretty, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(pretty))
}
