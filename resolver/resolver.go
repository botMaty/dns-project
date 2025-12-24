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
