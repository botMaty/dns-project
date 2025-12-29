package admin

import (
	"dns-server/types"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Server struct {
	store    types.Storage
	password string

	sessions map[string]time.Time
}

func New(store types.Storage, password string) *Server {
	return &Server{
		store:    store,
		password: password,
		sessions: make(map[string]time.Time),
	}
}

func (s *Server) isAdmin(r *http.Request) bool {
	c, err := r.Cookie("session")
	if err != nil {
		return false
	}

	exp, ok := s.sessions[c.Value]
	if !ok || time.Now().After(exp) {
		delete(s.sessions, c.Value) // پاکسازی session منقضی شده
		return false
	}

	return true
}

func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("/login", s.handleLogin)
	mux.HandleFunc("/logout", s.handleLogout)

	mux.HandleFunc("/", (s.handleUI))
	mux.HandleFunc("/admin/records", s.handleRecords)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Password != s.password {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token := uuid.NewString()
	exp := time.Now().Add(24 * time.Hour)
	s.sessions[token] = exp

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil, // فقط روی HTTPS secure باشد
		SameSite: http.SameSiteLaxMode,
		Expires:  exp,
	})

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session")
	if err == nil {
		delete(s.sessions, c.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

func (s *Server) handleRecords(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost || r.Method == http.MethodDelete {
		if !s.isAdmin(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

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
