# key.pem:
openssl genrsa -out key.pem 2048

# cert.pem:
openssl req -new -x509 -key key.pem -out cert.pem -days 365

# UDP test
dig @127.0.0.1 -p 8053 example.com

# TCP test
dig @127.0.0.1 -p 8053 example.com +tcp

# DoH GET test with cli with HTTP connection
go run cmd/doh-cli/main.go -name google.com

# DoH GET test with cli with HTTPs connection
go run cmd/doh-cli/main.go -name google.com -https true

# Web UI for handle records: http://127.0.0.1:8055

# DoH HTTP packet format:
GET /dns-query?dns=...
POST /dns-query
Response Type: binary
Content-Type: application/dns-message

# DNS over JSON (in HTTP)
GET /dns-query/json?name=google.com&type=A
POST /dns-query/json
{
  "name": "google.com",
  "type": "A"
}
Response Type: JSON:
{
  "rcode": "NOERROR",
  "answers": [
    {
      "name": "google.com.",
      "type": "A",
      "ttl": 300,
      "value": "142.250.185.78"
    }
  ]
}
Content-Type: application/dns-message