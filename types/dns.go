package types

import (
	"context"
	"time"
)

type RecordType uint16

type DNSQuestion struct {
	Name string
	Type RecordType
}

type DNSRecord struct {
	Name      string
	Type      RecordType
	Value     string
	TTL       uint32    // برای پاسخ
	ExpiresAt time.Time // برای منطق داخلی
}

type DNSResponse struct {
	Records []DNSRecord
	RCode   int
}

type Storage interface {
	Get(question DNSQuestion) ([]DNSRecord, bool)
	Set(record DNSRecord)
	Delete(name string, rtype RecordType)
}

type Resolver interface {
	Resolve(ctx context.Context, req []byte) ([]byte, error)
}

type UpStream interface {
	Query(question DNSQuestion) (DNSResponse, error)
}
