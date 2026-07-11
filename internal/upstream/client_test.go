package upstream

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientLogsInWithAgreementAndListsUsage(t *testing.T) {
	var settingsRead bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/settings/public":
			settingsRead = true
			writeTestJSON(w, map[string]any{"code": 0, "data": map[string]any{"login_agreement_revision": "revision-7"}})
		case "/api/v1/auth/login":
			if !settingsRead {
				t.Error("login occurred before public settings")
			}
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["email"] != "a@example.com" || body["password"] != "secret" || body["login_agreement_revision"] != "revision-7" {
				t.Errorf("unexpected login body %#v", body)
			}
			writeTestJSON(w, map[string]any{"code": 0, "data": map[string]any{"access_token": "access", "refresh_token": "refresh", "expires_in": 3600, "token_type": "Bearer"}})
		case "/api/v1/usage":
			if r.Header.Get("Authorization") != "Bearer access" {
				t.Errorf("authorization=%q", r.Header.Get("Authorization"))
			}
			q := r.URL.Query()
			if q.Get("page") != "2" || q.Get("page_size") != "50" || q.Get("timezone") != "Asia/Shanghai" || q.Get("sort_order") != "desc" {
				t.Errorf("unexpected query %v", q)
			}
			writeTestJSON(w, map[string]any{"code": 0, "data": map[string]any{"items": []any{map[string]any{"id": 9, "model": "gpt", "created_at": "2026-07-11T00:00:00Z"}}, "total": 1, "page": 2, "page_size": 50, "pages": 1}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "a@example.com", "secret", server.Client())
	if err := client.Login(context.Background()); err != nil {
		t.Fatal(err)
	}
	page, err := client.ListUsage(context.Background(), UsageQuery{Page: 2, PageSize: 50, StartTime: time.Now(), EndTime: time.Now(), Timezone: "Asia/Shanghai", SortOrder: "desc"})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != 9 || len(page.Items[0].Raw) == 0 {
		t.Fatalf("unexpected page %#v", page)
	}
}

func TestClientClassifiesRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "slow down", http.StatusTooManyRequests) }))
	defer server.Close()
	client := NewClient(server.URL, "a@example.com", "secret", server.Client())
	if err := client.Login(context.Background()); !IsKind(err, ErrorRateLimit) {
		t.Fatalf("error=%v", err)
	}
}

func writeTestJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(value)
}
