package syncer

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/ledger"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/upstream"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) UpsertAccount(ctx context.Context, id, name string, enabled bool) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := r.db.ExecContext(ctx, `INSERT INTO accounts(id,name,enabled,created_at,updated_at) VALUES(?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,enabled=excluded.enabled,updated_at=excluded.updated_at`, id, name, enabled, now, now)
	return err
}

func (r *Repository) RecordAccountSync(ctx context.Context, accountID, currentHost string, at time.Time, syncErr error) error {
	now := at.UTC().Format(time.RFC3339Nano)
	if syncErr == nil {
		_, err := r.db.ExecContext(ctx, `UPDATE accounts SET current_host=?,last_sync_at=?,last_error=NULL,updated_at=? WHERE id=?`, currentHost, now, now, accountID)
		return err
	}
	message := syncErr.Error()
	if len(message) > 500 {
		message = message[:500]
	}
	_, err := r.db.ExecContext(ctx, `UPDATE accounts SET current_host=?,last_error=?,updated_at=? WHERE id=?`, currentHost, message, now, accountID)
	return err
}

func (r *Repository) Cursor(ctx context.Context, accountID, dataType string) (time.Time, bool, error) {
	var raw string
	err := r.db.QueryRowContext(ctx, `SELECT cursor_at FROM sync_cursors WHERE account_id=? AND data_type=?`, accountID, dataType).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, err
	}
	value, err := time.Parse(time.RFC3339Nano, raw)
	return value, true, err
}

func (r *Repository) StoreUsage(ctx context.Context, accountID string, items []upstream.UsageRecord, cursor time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, item := range items {
		raw := item.Raw
		if !json.Valid(raw) {
			raw = []byte("{}")
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO usage_records(account_id,upstream_id,request_id,api_key_id,upstream_account_id,model,inbound_endpoint,upstream_endpoint,input_tokens,output_tokens,total_cost,actual_cost,duration_ms,first_token_ms,created_at,raw_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(account_id,upstream_id) DO UPDATE SET model=excluded.model,actual_cost=excluded.actual_cost,raw_json=excluded.raw_json`, accountID, item.ID, item.RequestID, item.APIKeyID, item.AccountID, item.Model, item.InboundEndpoint, item.UpstreamEndpoint, item.InputTokens, item.OutputTokens, number(item.TotalCost), number(item.ActualCost), item.DurationMS, item.FirstTokenMS, item.CreatedAt.UTC().Format(time.RFC3339Nano), string(raw))
		if err != nil {
			return fmt.Errorf("store usage %d: %w", item.ID, err)
		}
	}
	if err = setCursor(ctx, tx, accountID, "usage", cursor); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) StoreLedger(ctx context.Context, accountID string, items []upstream.LedgerEntry, cursor time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, item := range items {
		remark := ledger.Normalize(item.Reason, item.RefType, item.RefID, item.Metadata)
		extracted, _ := json.Marshal(remark.Fields)
		metadata := item.Metadata
		if !json.Valid(metadata) {
			metadata = []byte("{}")
		}
		raw := item.Raw
		if !json.Valid(raw) {
			raw = []byte("{}")
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO balance_ledger_entries(account_id,upstream_id,direction,amount,balance_after,reason,business_category,ref_type,ref_id,remark_text,search_text,extracted_json,metadata_json,created_at,raw_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(account_id,upstream_id) DO UPDATE SET amount=excluded.amount,balance_after=excluded.balance_after,remark_text=excluded.remark_text,search_text=excluded.search_text,metadata_json=excluded.metadata_json,raw_json=excluded.raw_json`, accountID, item.ID, item.Direction, number(item.Amount), number(item.BalanceAfter), item.Reason, remark.Category, item.RefType, item.RefID, remark.Text, remark.SearchText, string(extracted), string(metadata), item.CreatedAt.UTC().Format(time.RFC3339Nano), string(raw))
		if err != nil {
			return fmt.Errorf("store ledger %d: %w", item.ID, err)
		}
	}
	if err = setCursor(ctx, tx, accountID, "ledger", cursor); err != nil {
		return err
	}
	return tx.Commit()
}

func setCursor(ctx context.Context, tx *sql.Tx, accountID, dataType string, cursor time.Time) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO sync_cursors(account_id,data_type,cursor_at,updated_at) VALUES(?,?,?,?) ON CONFLICT(account_id,data_type) DO UPDATE SET cursor_at=excluded.cursor_at,updated_at=excluded.updated_at`, accountID, dataType, cursor.UTC().Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
func number(n json.Number) string {
	if n.String() == "" {
		return "0"
	}
	return n.String()
}
