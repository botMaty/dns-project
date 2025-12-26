package resolver

import (
	"context"
	"dns-server/types"
	"fmt"
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

type Resolver struct {
	store    types.Storage
	upstream types.UpStream
}

func NewResolver(store types.Storage, upstream types.UpStream) *Resolver {
	return &Resolver{
		store:    store,
		upstream: upstream,
	}
}

func (r *Resolver) Resolve(
	ctx context.Context,
	req []byte,
) ([]byte, error) {

	var p dnsmessage.Parser

	header, err := p.Start(req)
	if err != nil {
		return nil, err
	}

	q, err := p.Question()
	if err != nil {
		return nil, err
	}

	question := types.DNSQuestion{
		Name: q.Name.String(),
		Type: types.RecordType(q.Type),
	}

	if records, ok := r.store.Get(question); ok {
		return r.buildResponse(header, question, records)
	}

	resp, err := r.upstream.Query(question)
	if err != nil {
		return r.buildErrorResponse(header, dnsmessage.RCodeServerFailure)
	}

	for _, rec := range resp.Records {
		r.store.Set(rec)
	}

	return r.buildResponse(header, question, resp.Records)
}

func (r *Resolver) buildResponse(
	reqHeader dnsmessage.Header,
	q types.DNSQuestion,
	records []types.DNSRecord,
) ([]byte, error) {
	hdr := dnsmessage.Header{
		ID:                 reqHeader.ID,
		Response:           true,
		Authoritative:      true,
		RecursionAvailable: true,
		RCode:              dnsmessage.RCodeSuccess,
	}

	question := dnsmessage.Question{
		Name:  dnsmessage.MustNewName(q.Name),
		Type:  dnsmessage.Type(q.Type),
		Class: dnsmessage.ClassINET,
	}

	msg := dnsmessage.Message{
		Header:    hdr,
		Questions: []dnsmessage.Question{question},
	}
	for _, rec := range records {
		res, err := toResource(rec)
		if err != nil {
			continue
		}
		msg.Answers = append(msg.Answers, res)
	}

	return msg.Pack()
}

func toResource(rec types.DNSRecord) (dnsmessage.Resource, error) {
	h := dnsmessage.ResourceHeader{
		Name:  dnsmessage.MustNewName(rec.Name),
		Type:  dnsmessage.Type(rec.Type),
		Class: dnsmessage.ClassINET,
		TTL:   rec.TTL,
	}

	switch dnsmessage.Type(rec.Type) {
	case dnsmessage.TypeA:
		ip := net.ParseIP(rec.Value).To4()
		if ip == nil {
			return dnsmessage.Resource{}, fmt.Errorf("invalid A record")
		}
		return dnsmessage.Resource{
			Header: h,
			Body:   &dnsmessage.AResource{A: [4]byte(ip)},
		}, nil

	case dnsmessage.TypeAAAA:
		ip := net.ParseIP(rec.Value).To16()
		if ip == nil {
			return dnsmessage.Resource{}, fmt.Errorf("invalid AAAA record")
		}
		return dnsmessage.Resource{
			Header: h,
			Body:   &dnsmessage.AAAAResource{AAAA: [16]byte(ip)},
		}, nil

	case dnsmessage.TypeTXT:
		return dnsmessage.Resource{
			Header: h,
			Body:   &dnsmessage.TXTResource{TXT: []string{rec.Value}},
		}, nil

	case dnsmessage.TypeNS:
		return dnsmessage.Resource{
			Header: h,
			Body: &dnsmessage.NSResource{
				NS: dnsmessage.MustNewName(rec.Value),
			},
		}, nil

	case dnsmessage.TypePTR:
		return dnsmessage.Resource{
			Header: h,
			Body: &dnsmessage.PTRResource{
				PTR: dnsmessage.MustNewName(rec.Value),
			},
		}, nil

	default:
		return dnsmessage.Resource{}, fmt.Errorf("unsupported record type")
	}
}

func (r *Resolver) buildErrorResponse(
	reqHeader dnsmessage.Header,
	rcode dnsmessage.RCode,
) ([]byte, error) {
	hdr := dnsmessage.Header{
		ID:                 reqHeader.ID,
		Response:           true,
		Authoritative:      true,
		RecursionAvailable: true,
		RCode:              rcode,
	}

	msg := dnsmessage.Message{
		Header: hdr,
	}

	return msg.Pack()
}
