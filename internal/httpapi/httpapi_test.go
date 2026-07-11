package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/analytics"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/auth"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/secrets"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/store"
)

type testClock struct{ now time.Time }

func (c testClock) Now() time.Time { return c.now }

func TestRouterProtectsAnalyticsAndSetsSecureSession(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	clock := testClock{now: time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)}
	authService := auth.NewService(db, clock)
	if _, err = authService.SyncDashboardUser(context.Background(), secrets.DashboardCredentials{Username: "admin", Password: "password-123"}); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Dependencies{Auth: authService, Analytics: analytics.NewService(db, time.UTC), Clock: clock, PublicURL: "https://usage.example.com", SecureCookie: true})

	health := httptest.NewRecorder()
	router.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if health.Code != 200 {
		t.Fatalf("health=%d", health.Code)
	}
	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/overview?start=2026-07-10T00:00:00Z&end=2026-07-12T00:00:00Z&granularity=hour", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized=%d", unauthorized.Code)
	}

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "password-123"})
	login := httptest.NewRecorder()
	router.ServeHTTP(login, httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body)))
	if login.Code != 200 {
		t.Fatalf("login=%d body=%s", login.Code, login.Body.String())
	}
	result := login.Result()
	cookies := result.Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].MaxAge != 86400 {
		t.Fatalf("cookies=%#v", cookies)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/overview?start=2026-07-10T00:00:00Z&end=2026-07-12T00:00:00Z&granularity=hour", nil)
	req.AddCookie(cookies[0])
	authorized := httptest.NewRecorder()
	router.ServeHTTP(authorized, req)
	if authorized.Code != 200 {
		t.Fatalf("authorized=%d body=%s", authorized.Code, authorized.Body.String())
	}
	ledgerReq := httptest.NewRequest(http.MethodGet, "/api/ledger/entries?start=2026-07-10T00:00:00Z&end=2026-07-12T00:00:00Z&reason=usage_charge&remark=req", nil)
	ledgerReq.AddCookie(cookies[0])
	ledger := httptest.NewRecorder()
	router.ServeHTTP(ledger, ledgerReq)
	if ledger.Code != http.StatusOK {
		t.Fatalf("ledger=%d body=%s", ledger.Code, ledger.Body.String())
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.AddCookie(cookies[0])
	logoutReq.Header.Set("Origin", "https://evil.example")
	logout := httptest.NewRecorder()
	router.ServeHTTP(logout, logoutReq)
	if logout.Code != http.StatusForbidden {
		t.Fatalf("origin=%d", logout.Code)
	}
}
