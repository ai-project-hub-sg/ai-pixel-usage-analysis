package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	db       *sql.DB
	location *time.Location
}

func NewService(db *sql.DB, location *time.Location) *Service {
	return &Service{db: db, location: location}
}

func (s *Service) Overview(ctx context.Context, f Filters) (Overview, error) {
	if f.Granularity != "hour" && f.Granularity != "day" {
		return Overview{}, fmt.Errorf("invalid granularity")
	}
	buckets := map[time.Time]*Bucket{}
	accountClause, args := "", []any{f.Start.UTC().Format(time.RFC3339Nano), f.End.UTC().Format(time.RFC3339Nano)}
	if f.AccountID != "" {
		accountClause = " AND account_id=?"
		args = append(args, f.AccountID)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT created_at,input_tokens,output_tokens,actual_cost FROM usage_records WHERE created_at>=? AND created_at<?`+accountClause, args...)
	if err != nil {
		return Overview{}, err
	}
	for rows.Next() {
		var raw, cost string
		var input, output int64
		if err = rows.Scan(&raw, &input, &output, &cost); err != nil {
			rows.Close()
			return Overview{}, err
		}
		at, e := time.Parse(time.RFC3339Nano, raw)
		if e != nil {
			continue
		}
		key := s.bucket(at, f.Granularity)
		b := ensure(buckets, key)
		b.Requests++
		b.InputTokens += input
		b.OutputTokens += output
		b.ActualCost += decimal(cost)
	}
	rows.Close()
	rows, err = s.db.QueryContext(ctx, `SELECT created_at,direction,amount FROM balance_ledger_entries WHERE created_at>=? AND created_at<?`+accountClause, args...)
	if err != nil {
		return Overview{}, err
	}
	for rows.Next() {
		var raw, direction, amount string
		if err = rows.Scan(&raw, &direction, &amount); err != nil {
			rows.Close()
			return Overview{}, err
		}
		at, e := time.Parse(time.RFC3339Nano, raw)
		if e != nil {
			continue
		}
		b := ensure(buckets, s.bucket(at, f.Granularity))
		if direction == "credit" {
			b.Credit += decimal(amount)
		} else if direction == "debit" {
			b.Debit += decimal(amount)
		}
		b.Net = b.Credit - b.Debit
	}
	rows.Close()
	keys := make([]time.Time, 0, len(buckets))
	for key := range buckets {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Before(keys[j]) })
	result := Overview{Buckets: make([]Bucket, 0, len(keys))}
	for _, key := range keys {
		b := buckets[key]
		b.Yesterday = comparison(buckets[key.Add(-24*time.Hour)])
		b.LastWeek = comparison(buckets[key.Add(-7*24*time.Hour)])
		result.Buckets = append(result.Buckets, *b)
	}
	return result, nil
}
func (s *Service) bucket(value time.Time, granularity string) time.Time {
	local := value.In(s.location)
	if granularity == "day" {
		return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, s.location)
	}
	return time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), 0, 0, 0, s.location)
}
func ensure(values map[time.Time]*Bucket, key time.Time) *Bucket {
	if values[key] == nil {
		values[key] = &Bucket{Start: key}
	}
	return values[key]
}
func comparison(b *Bucket) *Comparison {
	if b == nil {
		return nil
	}
	return &Comparison{Requests: b.Requests, ActualCost: b.ActualCost, Credit: b.Credit, Debit: b.Debit, Net: b.Net}
}
func decimal(raw string) float64 { value, _ := strconv.ParseFloat(raw, 64); return value }

func (s *Service) UsageRecords(ctx context.Context, f Filters) ([]UsageRow, error) {
	query := `SELECT upstream_id,account_id,coalesce(model,''),coalesce(inbound_endpoint,''),coalesce(api_key_id,0),input_tokens,output_tokens,actual_cost,created_at FROM usage_records WHERE created_at>=? AND created_at<?`
	args := []any{f.Start.UTC().Format(time.RFC3339Nano), f.End.UTC().Format(time.RFC3339Nano)}
	if f.AccountID != "" {
		query += " AND account_id=?"
		args = append(args, f.AccountID)
	}
	if f.Model != "" {
		query += " AND model=?"
		args = append(args, f.Model)
	}
	if f.APIKeyID > 0 {
		query += " AND api_key_id=?"
		args = append(args, f.APIKeyID)
	}
	if f.Endpoint != "" {
		query += " AND inbound_endpoint=?"
		args = append(args, f.Endpoint)
	}
	rows, err := s.db.QueryContext(ctx, query+" ORDER BY created_at DESC LIMIT 500", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []UsageRow{}
	for rows.Next() {
		var row UsageRow
		if err = rows.Scan(&row.ID, &row.AccountID, &row.Model, &row.Endpoint, &row.APIKeyID, &row.InputTokens, &row.OutputTokens, &row.ActualCost, &row.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (s *Service) LedgerEntries(ctx context.Context, f Filters) ([]LedgerRow, error) {
	query := `SELECT upstream_id,account_id,direction,amount,balance_after,reason,business_category,remark_text,coalesce(ref_type,''),coalesce(ref_id,0),metadata_json,created_at FROM balance_ledger_entries WHERE created_at>=? AND created_at<?`
	args := []any{f.Start.UTC().Format(time.RFC3339Nano), f.End.UTC().Format(time.RFC3339Nano)}
	for _, item := range []struct{ column, value string }{{"account_id", f.AccountID}, {"reason", f.Reason}, {"business_category", f.Category}, {"direction", f.Direction}, {"ref_type", f.RefType}} {
		if item.value != "" {
			query += " AND " + item.column + "=?"
			args = append(args, item.value)
		}
	}
	if f.RefID > 0 {
		query += " AND ref_id=?"
		args = append(args, f.RefID)
	}
	if f.Remark != "" {
		query += ` AND lower(search_text) LIKE ? ESCAPE '\'`
		args = append(args, "%"+escapeLike(strings.ToLower(f.Remark))+"%")
	}
	rows, err := s.db.QueryContext(ctx, query+" ORDER BY created_at DESC LIMIT 500", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []LedgerRow{}
	for rows.Next() {
		var row LedgerRow
		var metadata string
		if err = rows.Scan(&row.ID, &row.AccountID, &row.Direction, &row.Amount, &row.BalanceAfter, &row.Reason, &row.Category, &row.Remark, &row.RefType, &row.RefID, &metadata, &row.CreatedAt); err != nil {
			return nil, err
		}
		row.Metadata = json.RawMessage(metadata)
		result = append(result, row)
	}
	return result, rows.Err()
}

func (s *Service) AccountStatuses(ctx context.Context) ([]AccountStatus, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,name,enabled,coalesce(current_host,''),coalesce(last_sync_at,''),coalesce(last_error,'') FROM accounts ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []AccountStatus{}
	for rows.Next() {
		var row AccountStatus
		if err = rows.Scan(&row.ID, &row.Name, &row.Enabled, &row.CurrentHost, &row.LastSyncAt, &row.LastError); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}
func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "%", `\%`)
	return strings.ReplaceAll(value, "_", `\_`)
}
