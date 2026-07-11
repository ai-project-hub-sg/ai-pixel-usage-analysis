package syncer

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/store"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/upstream"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type fakeUpstream struct {
	usageStart  time.Time
	ledgerStart time.Time
}

type failingUpstream struct{ fakeUpstream }

func (f *failingUpstream) ListUsage(context.Context, upstream.UsageQuery) (upstream.Page[upstream.UsageRecord], error) {
	return upstream.Page[upstream.UsageRecord]{}, errors.New("fixture upstream unavailable")
}

func (f *fakeUpstream) Login(context.Context) error   { return nil }
func (f *fakeUpstream) Refresh(context.Context) error { return nil }
func (f *fakeUpstream) ListUsage(_ context.Context, q upstream.UsageQuery) (upstream.Page[upstream.UsageRecord], error) {
	f.usageStart = q.StartTime
	return upstream.Page[upstream.UsageRecord]{Items: []upstream.UsageRecord{{ID: 1, Model: "gpt", CreatedAt: q.EndTime, Raw: json.RawMessage(`{"id":1}`)}}, Total: 1, Page: 1, PageSize: q.PageSize, Pages: 1}, nil
}
func (f *fakeUpstream) ListLedger(_ context.Context, q upstream.LedgerQuery) (upstream.Page[upstream.LedgerEntry], error) {
	f.ledgerStart = q.StartTime
	return upstream.Page[upstream.LedgerEntry]{Items: []upstream.LedgerEntry{{ID: 2, Direction: "debit", Amount: "1.25", BalanceAfter: "9.5", Reason: "usage_charge", Metadata: json.RawMessage(`{"request_id":"req-1"}`), CreatedAt: q.EndTime, Raw: json.RawMessage(`{"id":2}`)}}, Total: 1, Page: 1, PageSize: q.PageSize, Pages: 1}, nil
}

func TestInitialSyncStartsPreviousMonthAndIsIdempotent(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "sync.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewRepository(db)
	if err := repo.UpsertAccount(context.Background(), "primary", "主账户", true); err != nil {
		t.Fatal(err)
	}
	api := &fakeUpstream{}
	now := time.Date(2026, 7, 11, 5, 0, 0, 0, time.UTC)
	service := NewService(repo, map[string]upstream.API{"primary": api}, time.FixedZone("CST", 8*3600), fixedClock{now}, 5*time.Minute)
	if err := service.SyncAccount(context.Background(), "primary"); err != nil {
		t.Fatal(err)
	}
	if err := service.SyncAccount(context.Background(), "primary"); err != nil {
		t.Fatal(err)
	}
	wantStart := time.Date(2026, 5, 31, 16, 0, 0, 0, time.UTC)
	if api.usageStart.Before(wantStart) || api.ledgerStart.Before(wantStart) {
		t.Fatalf("starts usage=%v ledger=%v", api.usageStart, api.ledgerStart)
	}
	for table, want := range map[string]int{"usage_records": 1, "balance_ledger_entries": 1, "sync_cursors": 2} {
		var got int
		if err := db.QueryRow("SELECT count(*) FROM " + table).Scan(&got); err != nil || got != want {
			t.Fatalf("%s=%d err=%v", table, got, err)
		}
	}
}

func TestSyncAccountRecordsCurrentHostAndFreshness(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "status.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewRepository(db)
	if err = repo.UpsertAccount(context.Background(), "primary", "主账户", true); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 11, 5, 0, 0, 0, time.UTC)
	client := upstream.NewFailover([]upstream.Endpoint{{URL: "https://primary.fixture/", API: &fakeUpstream{}}})
	service := NewService(repo, map[string]upstream.API{"primary": client}, time.UTC, fixedClock{now}, 5*time.Minute)
	if err = service.SyncAccount(context.Background(), "primary"); err != nil {
		t.Fatal(err)
	}
	var host, lastSync, lastError string
	if err = db.QueryRow(`SELECT coalesce(current_host,''),coalesce(last_sync_at,''),coalesce(last_error,'') FROM accounts WHERE id='primary'`).Scan(&host, &lastSync, &lastError); err != nil {
		t.Fatal(err)
	}
	if host != "https://primary.fixture/" || lastSync != now.Format(time.RFC3339Nano) || lastError != "" {
		t.Fatalf("host=%q last_sync=%q last_error=%q", host, lastSync, lastError)
	}
}

func TestSyncAccountRecordsFailureWithoutAdvancingFreshness(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "failure.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewRepository(db)
	if err = repo.UpsertAccount(context.Background(), "primary", "主账户", true); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 11, 5, 0, 0, 0, time.UTC)
	client := upstream.NewFailover([]upstream.Endpoint{{URL: "https://failed.fixture/", API: &failingUpstream{}}})
	service := NewService(repo, map[string]upstream.API{"primary": client}, time.UTC, fixedClock{now}, 5*time.Minute)
	if err = service.SyncAccount(context.Background(), "primary"); err == nil {
		t.Fatal("expected sync failure")
	}
	var host, lastSync, lastError string
	if err = db.QueryRow(`SELECT coalesce(current_host,''),coalesce(last_sync_at,''),coalesce(last_error,'') FROM accounts WHERE id='primary'`).Scan(&host, &lastSync, &lastError); err != nil {
		t.Fatal(err)
	}
	if host != "https://failed.fixture/" || lastSync != "" || lastError == "" {
		t.Fatalf("host=%q last_sync=%q last_error=%q", host, lastSync, lastError)
	}
}
