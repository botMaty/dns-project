package resolver

import "dns-server/types"

type DefaultResolver struct {
	storage  types.Storage
	upstream types.UpStream
}

func NewResolver(s types.Storage, u types.UpStream) *DefaultResolver {
	return &DefaultResolver{
		storage:  s,
		upstream: u,
	}
}

func (r *DefaultResolver) Resolve(q types.DNSQuestion) types.DNSResponse {

	// 1. check local storage / cache
	if records, ok := r.storage.Get(q); ok {
		return types.DNSResponse{
			Records: records,
			RCode:   0,
		}
	}

	// 2. forward to upstream
	resp, err := r.upstream.Query(q)
	if err != nil {
		return types.DNSResponse{
			RCode: 2, // SERVFAIL
		}
	}

	// 3. cache upstream response
	for _, rec := range resp.Records {
		r.storage.Set(rec)
	}

	return resp
}
