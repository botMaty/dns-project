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