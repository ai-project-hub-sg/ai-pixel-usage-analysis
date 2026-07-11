package syncer

import (
	"context"
	"encoding/json"
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
