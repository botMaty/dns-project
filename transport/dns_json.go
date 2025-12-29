package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

func (s *DoHServer) handleJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var (
		name string
		typ  dnsmessage.Type
		err  error
	)

	// -------- GET --------
	switch r.Method {
	case http.MethodGet:
		name = r.URL.Query().Get("name")
		t := r.URL.Query().Get("type")
		if name == "" || t == "" {
			http.Error(w, "missing name or type", http.StatusBadRequest)
			return
		}
		typ, err = parseType(t)
		if err != nil {
			http.Error(w, "invalid type", http.StatusBadRequest)
			return
		}

	// -------- POST --------
	case http.MethodPost:
		if r.Header.Get("Content-Type") != "application/dns-json" {
			http.Error(w, "unsupported content-type", http.StatusUnsupportedMediaType)
			return
		}

		var req struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		name = req.Name
		typ, err = parseType(req.Type)
		if err != nil {
			http.Error(w, "invalid type", http.StatusBadRequest)
			return
		}
	}

	packet, err := buildDNSQuery(name, typ)
	if err != nil {
		http.Error(w, "dns build error", http.StatusInternalServerError)
		return
	}

	resp, err := s.resolver.Resolve(ctx, packet)
	if err != nil {
		http.Error(w, "resolver error", http.StatusInternalServerError)
		return
	}

	out, err := dnsToJSON(resp)
	if err != nil {
		http.Error(w, "parse error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/dns-json")
	json.NewEncoder(w).Encode(out)
}

func buildDNSQuery(name string, t dnsmessage.Type) ([]byte, error) {
	if !strings.HasSuffix(name, ".") {
		name += "."
	}

	msg := dnsmessage.Message{
		Header: dnsmessage.Header{ID: 1},
		Questions: []dnsmessage.Question{{
			Name:  dnsmessage.MustNewName(name),
			Type:  t,
			Class: dnsmessage.ClassINET,
		}},
	}
	return msg.Pack()
}

func parseType(t string) (dnsmessage.Type, error) {
	switch strings.ToUpper(t) {
	case "A":
		return dnsmessage.TypeA, nil
	case "AAAA":
		return dnsmessage.TypeAAAA, nil
	case "CNAME":
		return dnsmessage.TypeCNAME, nil
	case "MX":
		return dnsmessage.TypeMX, nil
	case "TXT":
		return dnsmessage.TypeTXT, nil
	case "NS":
		return dnsmessage.TypeNS, nil
	case "PTR":
		return dnsmessage.TypePTR, nil
	default:
		return 0, fmt.Errorf("unsupported type")
	}
}

func dnsToJSON(resp []byte) (map[string]any, error) {
	var p dnsmessage.Parser
	h, err := p.Start(resp)
	if err != nil {
		return nil, err
	}

	p.SkipAllQuestions()
	answers, _ := p.AllAnswers()

	out := []map[string]any{}
	for _, a := range answers {
		out = append(out, map[string]any{
			"name": a.Header.Name.String(),
			"type": a.Header.Type.String(),
			"ttl":  a.Header.TTL,
			"data": a.Body,
		})
	}

	return map[string]any{
		"rcode":   h.RCode.String(),
		"answers": out,
	}, nil
}
