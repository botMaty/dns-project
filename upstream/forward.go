package upstream

import (
	"crypto/rand"
	"dns-server/types"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

type UDPUpstream struct {
	server  string
	timeout time.Duration
}

func NewUDPUpstream(server string) *UDPUpstream {
	return &UDPUpstream{
		server:  server,
		timeout: 3 * time.Second,
	}
}

func toDNSMessageQuestion(q types.DNSQuestion) dnsmessage.Question {
	return dnsmessage.Question{
		Name:  dnsmessage.MustNewName(q.Name),
		Type:  dnsmessage.Type(q.Type),
		Class: dnsmessage.ClassINET,
	}
}

func buildQueryPacket(q types.DNSQuestion) ([]byte, uint16, error) {

	id, err := rand.Int(rand.Reader, big.NewInt(65535))
	if err != nil {
		return nil, 0, err
	}

	msg := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:       uint16(id.Int64()),
			Response: false,
		},
		Questions: []dnsmessage.Question{
			toDNSMessageQuestion(q),
		},
	}

	buf, err := msg.Pack()
	if err != nil {
		return nil, 0, err
	}

	return buf, msg.Header.ID, nil
}

func (u *UDPUpstream) exchange(packet []byte) ([]byte, error) {
	conn, err := net.DialTimeout("udp", u.server, u.timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(u.timeout))

	if _, err := conn.Write(packet); err != nil {
		return nil, err
	}

	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

func parseResponse(buf []byte) (types.DNSResponse, error) {
	var p dnsmessage.Parser

	hdr, err := p.Start(buf)
	if err != nil {
		return types.DNSResponse{}, err
	}

	if err := p.SkipAllQuestions(); err != nil {
		return types.DNSResponse{}, err
	}

	answers, err := p.AllAnswers()
	if err != nil {
		return types.DNSResponse{}, err
	}

	resp := types.DNSResponse{
		RCode: int(hdr.RCode),
	}

	for _, ans := range answers {
		rec, ok := convertAnswer(ans)
		if ok {
			resp.Records = append(resp.Records, rec)
		}
	}

	return resp, nil
}

func convertAnswer(a dnsmessage.Resource) (types.DNSRecord, bool) {
	rec := types.DNSRecord{
		Name: a.Header.Name.String(),
		Type: types.RecordType(a.Header.Type),
		TTL:  a.Header.TTL,
	}

	switch body := a.Body.(type) {
	case *dnsmessage.AResource:
		rec.Value = net.IP(body.A[:]).String()

	case *dnsmessage.AAAAResource:
		rec.Value = net.IP(body.AAAA[:]).String()

	case *dnsmessage.CNAMEResource:
		rec.Value = body.CNAME.String()

	case *dnsmessage.MXResource:
		// ذخیره MX به فرمت: "preference mx-server"
		rec.Value = fmt.Sprintf("%d %s", body.Pref, body.MX.String())

	case *dnsmessage.TXTResource:
		rec.Value = strings.Join(body.TXT, " ")

	case *dnsmessage.NSResource:
		rec.Value = body.NS.String()

	case *dnsmessage.PTRResource:
		rec.Value = body.PTR.String()

	default:
		return types.DNSRecord{}, false
	}

	return rec, true
}

func (u *UDPUpstream) Query(q types.DNSQuestion) (types.DNSResponse, error) {

	packet, _, err := buildQueryPacket(q)
	if err != nil {
		return types.DNSResponse{}, err
	}

	respBuf, err := u.exchange(packet)
	if err != nil {
		return types.DNSResponse{}, err
	}

	return parseResponse(respBuf)
}
