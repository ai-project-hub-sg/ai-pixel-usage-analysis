package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/auth"
)

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input) != nil {
		writeError(w, 400, "invalid_request", "invalid request")
		return
	}
	key := loginLimitKey(r, input.Username)
	now := s.deps.Clock.Now()
	if s.loginLimiter.blocked(key, now) {
		w.Header().Set("Retry-After", "900")
		writeError(w, http.StatusTooManyRequests, "login_rate_limited", "too many login attempts")
		return
	}
	session, err := s.deps.Auth.Login(r.Context(), input.Username, input.Password)
	if err != nil {
		s.loginLimiter.fail(key, now)
		writeError(w, 401, "invalid_credentials", "invalid username or password")
		return
	}
	s.loginLimiter.reset(key)
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: session.Token, Path: "/", Expires: session.ExpiresAt, MaxAge: int((24 * time.Hour).Seconds()), HttpOnly: true, Secure: s.deps.SecureCookie, SameSite: http.SameSiteLaxMode})
	writeJSON(w, 200, map[string]any{"expires_at": session.ExpiresAt})
}
func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie(sessionCookie)
	if cookie != nil {
		s.deps.Auth.Logout(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Path: "/", MaxAge: -1, HttpOnly: true, Secure: s.deps.SecureCookie, SameSite: http.SameSiteLaxMode})
	writeJSON(w, 200, map[string]bool{"logged_out": true})
}
func (s *server) session(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(userContextKey{}).(auth.User)
	writeJSON(w, 200, map[string]any{"username": user.Username})
}
