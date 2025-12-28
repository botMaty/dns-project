package storage

import (
	"dns-server/types"
	"sync"
	"time"
)

type MemoryStorage struct {
	mu      sync.RWMutex
	records map[string][]types.DNSRecord
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		records: make(map[string][]types.DNSRecord),
	}
}

func key(name string, rtype types.RecordType) string {
	return name + ":" + string(rune(rtype))
}

func (m *MemoryStorage) Get(q types.DNSQuestion) ([]types.DNSRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	k := key(q.Name, q.Type)
	recs, ok := m.records[k]
	if !ok {
		return nil, false
	}

	now := time.Now()
	valid := make([]types.DNSRecord, 0)

	for _, r := range recs {
		if r.ExpiresAt.After(now) {
			valid = append(valid, r)
		}
	}

	if len(valid) == 0 {
		return nil, false
	}

	return valid, true
}

func (m *MemoryStorage) Set(r types.DNSRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()

	r.ExpiresAt = time.Now().Add(time.Duration(r.TTL) * time.Second)
	k := key(r.Name, r.Type)

	m.records[k] = append(m.records[k], r)
}

func (m *MemoryStorage) Delete(name string, rtype types.RecordType, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	k := key(name, rtype)
	if value == "" {
		// delete all
		delete(m.records, k)
	} else {
		var filtered []types.DNSRecord
		for _, r := range m.records[k] {
			if r.Value != value {
				filtered = append(filtered, r)
			}
		}
		m.records[k] = filtered
	}
}

func (m *MemoryStorage) List() []types.DNSRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var all []types.DNSRecord
	for _, recs := range m.records {
		for _, r := range recs {
			if r.ExpiresAt.After(time.Now()) {
				all = append(all, r)
			}
		}
	}
	return all
}
