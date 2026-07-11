package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
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
