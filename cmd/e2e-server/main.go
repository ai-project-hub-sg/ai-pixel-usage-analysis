// Command e2e-server starts an isolated deterministic application instance for Playwright.
// It never reads config.toml or .env and never constructs a real upstream client.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/analytics"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/auth"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/httpapi"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/secrets"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/store"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/syncer"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/upstream"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/webui"
)

var fixtureNow = time.Date(2026, 7, 11, 7, 0, 0, 0, time.UTC)

type fixedClock struct{ now time.Time }

func (clock fixedClock) Now() time.Time { return clock.now }

func main() {
	address := flag.String("addr", "127.0.0.1:18080", "fixture server address")
	flag.Parse()
	if err := run(*address); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(address string) error {
	directory, err := os.MkdirTemp("", "ai-pixel-e2e-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(directory)
	db, err := store.Open(filepath.Join(directory, "fixtures.db"))
	if err != nil {
		return err
	}
	defer db.Close()
	ctx := context.Background()
	repository := syncer.NewRepository(db)
	if err = seed(ctx, db, repository); err != nil {
		return err
	}
	clock := fixedClock{now: fixtureNow}
	authService := auth.NewService(db, clock)
	if _, err = authService.SyncDashboardUser(ctx, secrets.DashboardCredentials{Username: "e2e-admin", Password: "e2e-password"}); err != nil {
		return err
	}
	origin := "http://" + address
	trigger := func(ctx context.Context, accountID string) error {
		result, updateErr := db.ExecContext(ctx, `UPDATE accounts SET last_sync_at=?,last_error=NULL,updated_at=? WHERE id=?`, fixtureNow.Format(time.RFC3339Nano), fixtureNow.Format(time.RFC3339Nano), accountID)
		if updateErr != nil {
			return updateErr
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return fmt.Errorf("unknown account %q", accountID)
		}
		return nil
	}
	router := httpapi.NewRouter(httpapi.Dependencies{Auth: authService, Analytics: analytics.NewService(db, time.FixedZone("CST", 8*3600)), Clock: clock, PublicURL: origin, Static: webui.Handler(), TriggerSync: trigger})
	server := &http.Server{Addr: address, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	shutdown, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-shutdown.Done()
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(closeCtx)
	}()
	err = server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func seed(ctx context.Context, db *sql.DB, repository *syncer.Repository) error {
	for _, account := range []struct{ id, name string }{{"primary", "主账户"}, {"secondary", "备用账户"}} {
		if err := repository.UpsertAccount(ctx, account.id, account.name, true); err != nil {
			return err
		}
	}
	if _, err := db.ExecContext(ctx, `UPDATE accounts SET current_host=?,last_sync_at=?,last_error=NULL WHERE id='primary'`, "https://primary.fixture", fixtureNow.Add(-2*time.Minute).Format(time.RFC3339Nano)); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `UPDATE accounts SET current_host=?,last_sync_at=?,last_error=? WHERE id='secondary'`, "https://secondary.fixture", fixtureNow.Add(-35*time.Minute).Format(time.RFC3339Nano), "上游连接超时"); err != nil {
		return err
	}
	primaryUsage := []upstream.UsageRecord{
		usage(1, "gpt-4.1", "/v1/responses", 101, fixtureNow.Add(-time.Hour), "0.42"),
		usage(2, "gpt-4.1", "/v1/responses", 101, fixtureNow.Add(-25*time.Hour), "0.31"),
		usage(3, "gpt-4.1", "/v1/responses", 101, fixtureNow.Add(-7*24*time.Hour-time.Hour), "0.27"),
	}
	if err := repository.StoreUsage(ctx, "primary", primaryUsage, fixtureNow); err != nil {
		return err
	}
	if err := repository.StoreUsage(ctx, "secondary", []upstream.UsageRecord{usage(11, "claude-sonnet", "/v1/messages", 202, fixtureNow.Add(-2*time.Hour), "0.18")}, fixtureNow); err != nil {
		return err
	}
	primaryLedger := []upstream.LedgerEntry{
		{ID: 100, Direction: "debit", Amount: json.Number("1.25"), BalanceAfter: json.Number("98.75"), Reason: "usage_charge", RefType: "request", RefID: 9001, Metadata: json.RawMessage(`{"request_id":"req-001","model":"gpt-4.1","api_key_id":101}`), CreatedAt: fixtureNow.Add(-45 * time.Minute), Raw: json.RawMessage(`{"id":100}`)},
		{ID: 101, Direction: "credit", Amount: json.Number("20"), BalanceAfter: json.Number("100"), Reason: "redeem_code", RefType: "code", RefID: 7001, Metadata: json.RawMessage(`{"code":"WELCOME"}`), CreatedAt: fixtureNow.Add(-3 * time.Hour), Raw: json.RawMessage(`{"id":101}`)},
	}
	if err := repository.StoreLedger(ctx, "primary", primaryLedger, fixtureNow); err != nil {
		return err
	}
	secondary := upstream.LedgerEntry{ID: 110, Direction: "debit", Amount: json.Number("0.5"), BalanceAfter: json.Number("49.5"), Reason: "usage_charge", Metadata: json.RawMessage(`{"request_id":"req-secondary"}`), CreatedAt: fixtureNow.Add(-2 * time.Hour), Raw: json.RawMessage(`{"id":110}`)}
	return repository.StoreLedger(ctx, "secondary", []upstream.LedgerEntry{secondary}, fixtureNow)
}

func usage(id int64, model, endpoint string, apiKeyID int64, at time.Time, cost string) upstream.UsageRecord {
	return upstream.UsageRecord{ID: id, RequestID: fmt.Sprintf("req-%03d", id), APIKeyID: apiKeyID, Model: model, InboundEndpoint: endpoint, InputTokens: 120, OutputTokens: 48, TotalCost: json.Number(cost), ActualCost: json.Number(cost), DurationMS: 620, FirstTokenMS: 190, CreatedAt: at, Raw: json.RawMessage(fmt.Sprintf(`{"id":%d}`, id))}
}
