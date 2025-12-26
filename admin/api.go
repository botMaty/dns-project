package admin

import (
	"dns-server/types"
	"encoding/json"
	"net/http"
)

type Server struct {
	store types.Storage
}

func New(store types.Storage) *Server {
	return &Server{store: store}
}

func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("/admin/records", s.handleRecords)
}

func (s *Server) handleRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

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
			Name string           `json:"name"`
			Type types.RecordType `json:"type"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		s.store.Delete(req.Name, req.Type)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
