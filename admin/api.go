package admin

import (
	"dns-server/types"
	"encoding/json"
	"net/http"
)

type Server struct {
	store types.Storage
	token string
}

func New(store types.Storage, token string) *Server {
	return &Server{
		store: store,
		token: token,
	}
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h != "Bearer "+s.token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("/", s.handleUI)
	mux.HandleFunc("/admin/records", s.auth(s.handleRecords))
}

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

func (s *Server) handleRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		records := s.store.List()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)

	case http.MethodPost:
		var req struct {
			Name  string           `json:"name"`
			Type  types.RecordType `json:"type"`
			Value string           `json:"value"`
			TTL   uint32           `json:"ttl"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		rec := types.DNSRecord{
			Name:  req.Name,
			Type:  req.Type,
			Value: req.Value,
			TTL:   req.TTL,
		}

		s.store.Set(rec)
		w.WriteHeader(http.StatusCreated)

	case http.MethodDelete:
		var req struct {
			Name  string           `json:"name"`
			Type  types.RecordType `json:"type"`
			Value string           `json:"value"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		s.store.Delete(req.Name, req.Type, req.Value)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
