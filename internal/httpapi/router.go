package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/analytics"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/auth"
)

const sessionCookie = "ai_pixel_session"

type Clock interface{ Now() time.Time }
type Dependencies struct {
	Auth         *auth.Service
	Analytics    *analytics.Service
	Clock        Clock
	PublicURL    string
	SecureCookie bool
	Static       http.Handler
	TriggerSync  func(context.Context, string) error
}
type server struct {
	deps         Dependencies
	publicOrigin string
}

func NewRouter(deps Dependencies) http.Handler {
	u, _ := url.Parse(deps.PublicURL)
	s := &server{deps: deps}
	if u != nil {
		s.publicOrigin = u.Scheme + "://" + u.Host
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]any{"status": "ok"}) })
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]any{"status": "ready"}) })
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.Handle("POST /api/auth/logout", s.protected(http.HandlerFunc(s.logout), true))
	mux.Handle("GET /api/auth/session", s.protected(http.HandlerFunc(s.session), false))
	mux.Handle("GET /api/overview", s.protected(http.HandlerFunc(s.overview), false))
	mux.Handle("GET /api/usage/records", s.protected(http.HandlerFunc(s.usageRecords), false))
	mux.Handle("GET /api/ledger/entries", s.protected(http.HandlerFunc(s.ledgerEntries), false))
	mux.Handle("GET /api/accounts/status", s.protected(http.HandlerFunc(s.accountStatuses), false))
	mux.Handle("POST /api/accounts/{id}/sync", s.protected(http.HandlerFunc(s.triggerSync), true))
	if deps.Static != nil {
		mux.Handle("GET /", deps.Static)
	}
	return mux
}

type userContextKey struct{}

func (s *server) protected(next http.Handler, write bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if write && s.publicOrigin != "" && r.Header.Get("Origin") != s.publicOrigin {
			writeError(w, http.StatusForbidden, "origin_forbidden", "origin is not allowed")
			return
		}
		cookie, err := r.Cookie(sessionCookie)
		if err != nil {
			writeError(w, 401, "unauthenticated", "authentication required")
			return
		}
		user, err := s.deps.Auth.Authenticate(r.Context(), cookie.Value)
		if err != nil {
			writeError(w, 401, "unauthenticated", "authentication required")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey{}, user)))
	})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"data": value})
}
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"code": code, "message": message}})
}
