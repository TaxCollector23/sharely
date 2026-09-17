package server

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/TaxCollector23/sharely/internal/qr"
	"github.com/TaxCollector23/sharely/internal/sharing"
)

// APIHandler implements the loopback-only control plane: the dashboard UI
// and other `sharely` CLI invocations both talk to it. It is never reachable
// from the LAN, which is what keeps a malicious shared page from ever being
// able to call it (see internal/server/content.go's isolation note).
type APIHandler struct {
	Manager     *sharing.Manager
	ContentHost string // LAN host:port shown in share URLs
	LocalName   string // "sharely.local" if advertised, else ""
	Interface   string // e.g. "Wi-Fi (en0)"
	StartedAt   time.Time
	Version     string
	Shutdown    func()
}

type shareDTO struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Target        string  `json:"target"`
	Type          string  `json:"type"`
	Status        string  `json:"status"`
	URL           string  `json:"url"`
	PrimaryURL    string  `json:"primaryUrl"`
	Remaining     string  `json:"remaining"`
	Duration      string  `json:"duration"`
	CreatedAt     string  `json:"createdAt"`
	ExpiresAt     *string `json:"expiresAt"`
	HasPassword   bool    `json:"hasPassword"`
	DeviceCount   int     `json:"deviceCount"`
	LastAccess    *string `json:"lastAccess"`
	NetworkAddr   string  `json:"networkAddress"`
	LocalHostname string  `json:"localHostname"`
}

func (h *APIHandler) toDTO(s *sharing.Share) shareDTO {
	var exp *string
	if s.ExpiresAt != nil {
		v := s.ExpiresAt.Format(time.RFC3339)
		exp = &v
	}
	var last *string
	if la := s.LastAccess(); la != nil {
		v := la.Format(time.RFC3339)
		last = &v
	}
	path := "/" + url.PathEscape(s.ID) + "/"
	lanURL := "http://" + h.ContentHost + path
	primary := lanURL
	if h.LocalName != "" {
		localHost := h.LocalName
		if _, port, err := net.SplitHostPort(h.ContentHost); err == nil {
			localHost = net.JoinHostPort(h.LocalName, port)
		}
		primary = "http://" + localHost + path
	}
	return shareDTO{
		ID:            s.ID,
		Name:          filepath.Base(s.TargetArg),
		Target:        s.TargetArg,
		Type:          string(s.Type),
		Status:        string(s.Status()),
		URL:           lanURL,
		PrimaryURL:    primary,
		Remaining:     s.RemainingLabel(),
		Duration:      string(s.Duration),
		CreatedAt:     s.CreatedAt.Format(time.RFC3339),
		ExpiresAt:     exp,
		HasPassword:   s.HasPassword,
		DeviceCount:   s.DeviceCount(),
		LastAccess:    last,
		NetworkAddr:   h.ContentHost,
		LocalHostname: h.LocalName,
	}
}

func (h *APIHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/shares", h.handleShares)
	mux.HandleFunc("/api/shares/", h.handleShareByID)
	mux.HandleFunc("/api/status", h.handleStatus)
	mux.HandleFunc("/api/shutdown", h.handleShutdown)
	return mux
}

func (h *APIHandler) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Requiring JSON makes a cross-origin HTML form unable to stop the local
	// service; scripted cross-origin requests are blocked by the browser's
	// preflight because this control server never enables CORS.
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "content type must be application/json", http.StatusUnsupportedMediaType)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"stopping": true})
	if h.Shutdown != nil {
		go h.Shutdown()
	}
}

func (h *APIHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version":     h.Version,
		"startedAt":   h.StartedAt.Format(time.RFC3339),
		"network":     h.ContentHost,
		"localHost":   h.LocalName,
		"interface":   h.Interface,
		"activeCount": len(h.Manager.Active()),
	})
}

func (h *APIHandler) handleShares(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		out := []shareDTO{}
		for _, s := range h.Manager.List() {
			out = append(out, h.toDTO(s))
		}
		writeJSON(w, http.StatusOK, out)
	case http.MethodPost:
		h.handleCreate(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type createRequest struct {
	Target   string `json:"target"`
	Duration string `json:"duration"`
	Password string `json:"password"`
}

func (h *APIHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	resolved, err := sharing.Resolve(req.Target)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	dur := sharing.Duration(req.Duration)
	if dur == "" {
		dur = sharing.Duration1Hour
	}
	s, err := h.Manager.Create(sharing.CreateOptions{
		TargetArg:     req.Target,
		RootDir:       resolved.RootDir,
		EntryPoint:    resolved.EntryPoint,
		Type:          resolved.Type,
		ProxyUpstream: resolved.ProxyUpstream,
		Duration:      dur,
		Password:      req.Password,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, h.toDTO(s))
}

func (h *APIHandler) handleShareByID(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/shares/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	id := parts[0]
	s, ok := h.Manager.Get(id)
	if !ok {
		http.Error(w, "no active share named \""+id+"\"", http.StatusNotFound)
		return
	}

	if len(parts) == 1 {
		if r.Method == http.MethodDelete {
			h.Manager.Stop(id)
			writeJSON(w, http.StatusOK, h.toDTO(s))
			return
		}
		writeJSON(w, http.StatusOK, h.toDTO(s))
		return
	}

	switch parts[1] {
	case "stop":
		h.Manager.Stop(id)
		writeJSON(w, http.StatusOK, h.toDTO(s))
	case "expiration":
		var body struct {
			Duration string `json:"duration"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		s.SetExpiration(sharing.Duration(body.Duration))
		writeJSON(w, http.StatusOK, h.toDTO(s))
	case "qr.svg":
		path := "/" + s.ID + "/"
		target := "http://" + h.ContentHost + path
		if h.LocalName != "" {
			target = "http://" + h.LocalName + path
		}
		svg, err := qr.SVG(target, 480)
		if err != nil {
			http.Error(w, "could not generate QR code", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write([]byte(svg))
	case "logs":
		writeJSON(w, http.StatusOK, s.Logs())
	default:
		http.NotFound(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
