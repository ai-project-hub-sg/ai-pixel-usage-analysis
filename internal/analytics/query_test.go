package analytics

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/store"
)

func TestOverviewSeparatesAccountsDirectionsAndComparisons(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "analytics.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	now := time.Date(2026, 7, 11, 2, 0, 0, 0, time.UTC)
	for _, id := range []string{"a", "b"} {
		_, err = db.Exec(`INSERT INTO accounts(id,name,enabled,created_at,updated_at) VALUES(?,?,1,?,?)`, id, id, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
		if err != nil {
			t.Fatal(err)
		}
	}
	insertUsage := func(account string, id int, at time.Time, cost string) {
		_, err = db.Exec(`INSERT INTO usage_records(account_id,upstream_id,model,input_tokens,output_tokens,total_cost,actual_cost,created_at,raw_json) VALUES(?,?,?,?,?,?,?,?,?)`, account, id, "gpt", 10, 5, cost, cost, at.Format(time.RFC3339Nano), `{}`)
		if err != nil {
			t.Fatal(err)
		}
	}
	insertUsage("a", 1, now, "2.5")
	insertUsage("b", 2, now, "1.5")
	insertUsage("a", 3, now.Add(-24*time.Hour), "1")
	insertUsage("a", 4, now.Add(-7*24*time.Hour), "2")
	for i, direction := range []string{"credit", "debit"} {
		_, err = db.Exec(`INSERT INTO balance_ledger_entries(account_id,upstream_id,direction,amount,balance_after,reason,business_category,remark_text,search_text,extracted_json,metadata_json,created_at,raw_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, "a", 10+i, direction, []string{"5", "3"}[i], "10", "usage_charge", "usage", "remark", "remark", `{}`, `{}`, now.Format(time.RFC3339Nano), `{}`)
		if err != nil {
			t.Fatal(err)
		}
	}
	svc := NewService(db, time.UTC)
	result, err := svc.Overview(ctx, Filters{Start: now.Add(-8 * 24 * time.Hour), End: now.Add(time.Hour), Granularity: "hour"})
	if err != nil {
		t.Fatal(err)
	}
	current := result.Buckets[len(result.Buckets)-1]
	if current.Requests != 2 || current.ActualCost != 4 || current.Credit != 5 || current.Debit != 3 || current.Net != 2 {
		t.Fatalf("current=%#v", current)
	}
	if current.Yesterday == nil || current.Yesterday.Requests != 1 || current.LastWeek == nil || current.LastWeek.Requests != 1 {
		t.Fatalf("comparisons=%#v", current)
	}
	single, err := svc.Overview(ctx, Filters{AccountID: "b", Start: now.Add(-time.Hour), End: now.Add(time.Hour), Granularity: "hour"})
	if err != nil {
		t.Fatal(err)
	}
	if len(single.Buckets) != 1 || single.Buckets[0].ActualCost != 1.5 {
		t.Fatalf("single=%#v", single)
	}
}
