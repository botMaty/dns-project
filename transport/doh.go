package transport

import (
	"context"
	"dns-server/types"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"time"
)

type DoHServer struct {
	addr     string
	resolver types.Resolver
	certFile string
	keyFile  string
}

func NewDoHServer(
	addr string,
	r types.Resolver,
	certFile string,
	keyFile string,
) *DoHServer {
	return &DoHServer{
		addr:     addr,
		resolver: r,
		certFile: certFile,
		keyFile:  keyFile,
	}
}

func (s *DoHServer) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/dns-query", s.handleDNS)
	mux.HandleFunc("/dns-query/json", s.handleJSON)

	server := &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// HTTP or HTTPS
	if s.certFile != "" && s.keyFile != "" {
		if _, err := os.Stat(s.certFile); err == nil {
			if _, err := os.Stat(s.keyFile); err == nil {
				return server.ListenAndServeTLS(s.certFile, s.keyFile)
			}
		}
	}
	return server.ListenAndServe()

}

func (s *DoHServer) handleDNS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req []byte
	var err error

	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query().Get("dns")
		if q == "" {
			http.Error(w, "Missing dns parameter", http.StatusBadRequest)
			return
		}

		req, err = base64.RawURLEncoding.DecodeString(q)
		if err != nil {
			http.Error(w, "Invalid base64", http.StatusBadRequest)
			return
		}

	case http.MethodPost:
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		defer r.Body.Close()

		req, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
	}

	resp, err := s.resolver.Resolve(ctx, req)
	if err != nil {
		http.Error(w, "Resolver error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/dns-message")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
