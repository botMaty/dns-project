# DNS Project

This project is a full-featured DNS server supporting:

* UDP and TCP on port 8053
* DNS over HTTPS (DoH) on port 8054
* Web UI for managing records on port 8055
* DNS over JSON API for querying records

## TLS Setup (for HTTPS)

For testing with a self-signed certificate:

```bash
mkdir certs
openssl genrsa -out certs/key.pem 2048
openssl req -new -x509 -key certs/key.pem -out certs/cert.pem -days 365
```

## Ports

| Service            | Port |
| ------------------ | ---- |
| UDP/TCP DNS        | 8053 |
| DoH (HTTPS)        | 8054 |
| Web UI / Admin API | 8055 |

## Testing

### UDP

```bash
dig @127.0.0.1 -p 8053 example.com
```

### TCP

```bash
dig @127.0.0.1 -p 8053 example.com +tcp
```

### DoH CLI

```bash
# GET request
go run cmd/doh-cli/main.go -name google.com

# Specify the Type
go run cmd/doh-cli/main.go -name google.com -type a

# POST request
go run cmd/doh-cli/main.go -name google.com -method post

# HTTPS
go run cmd/doh-cli/main.go -name google.com -https true
```

### DNS JSON CLI

```bash
# GET request
go run cmd/dns-json-cli/main.go -name google.com

# Specify the Type
go run cmd/dns-json-cli/main.go -name google.com -type a

# POST request
go run cmd/dns-json-cli/main.go -name google.com -method post

# HTTPS
go run cmd/dns-json-cli/main.go -name google.com -https true
```

### Web UI

```
http://127.0.0.1:8055
```

## DNS Request Formats

### DoH (binary)

* GET: `/dns-query?dns=...`
* POST: `/dns-query` (body: DNS binary)
* Response: `Content-Type: application/dns-message`

### DNS over JSON

* GET: `/dns-query/json?name=google.com&type=A`
* POST: `/dns-query/json`

```json
{ "name": "google.com", "type": "A" }
```

* Response: `Content-Type: application/dns-json`

```json
{
  "rcode": "NOERROR",
  "answers": [
    { "name": "google.com.", "type": "A", "ttl": 300, "value": "142.250.185.78" }
  ]
}
```

## Sample `.env`

```env
UDP_PORT=:8053
TCP_PORT=:8053
DOH_PORT=:8054
ADMIN_PORT=:8055

DOH_CERT=certs/cert.pem
DOH_KEY=certs/key.pem

ADMIN_HASHED_PASSWORD='$2a$10$rKkwknuEbrrudD5TsW8sjOZlLAfEioBgqKLIpCYJjLwq1vtNHUDKm'
UPSTREAM_DNS=8.8.8.8:53
DB_FILE=dns_records.db
```

## Run the Project

```bash
go run cmd/main.go
```
