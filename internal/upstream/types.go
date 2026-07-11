package upstream

import (
	"context"
	"encoding/json"
	"time"
)

type Page[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Pages    int `json:"pages"`
}

type UsageRecord struct {
	ID               int64           `json:"id"`
	RequestID        string          `json:"request_id"`
	APIKeyID         int64           `json:"api_key_id"`
	AccountID        int64           `json:"account_id"`
	Model            string          `json:"model"`
	InboundEndpoint  string          `json:"inbound_endpoint"`
	UpstreamEndpoint string          `json:"upstream_endpoint"`
	InputTokens      int64           `json:"input_tokens"`
	OutputTokens     int64           `json:"output_tokens"`
	TotalCost        json.Number     `json:"total_cost"`
	ActualCost       json.Number     `json:"actual_cost"`
	DurationMS       int64           `json:"duration_ms"`
	FirstTokenMS     int64           `json:"first_token_ms"`
	CreatedAt        time.Time       `json:"created_at"`
	Raw              json.RawMessage `json:"-"`
}

func (r *UsageRecord) UnmarshalJSON(data []byte) error {
	type plain UsageRecord
	if err := json.Unmarshal(data, (*plain)(r)); err != nil {
		return err
	}
	r.Raw = append(r.Raw[:0], data...)
	return nil
}

type LedgerEntry struct {
	ID           int64           `json:"id"`
	Direction    string          `json:"direction"`
	Amount       json.Number     `json:"amount"`
	Reason       string          `json:"reason"`
	RefType      string          `json:"ref_type"`
	RefID        int64           `json:"ref_id"`
	BalanceAfter json.Number     `json:"balance_after"`
	Metadata     json.RawMessage `json:"metadata"`
	CreatedAt    time.Time       `json:"created_at"`
	Raw          json.RawMessage `json:"-"`
}

func (r *LedgerEntry) UnmarshalJSON(data []byte) error {
	type plain LedgerEntry
	if err := json.Unmarshal(data, (*plain)(r)); err != nil {
		return err
	}
	r.Raw = append(r.Raw[:0], data...)
	return nil
}

type UsageQuery struct {
	Page, PageSize      int
	StartTime, EndTime  time.Time
	Timezone, SortOrder string
}
type LedgerQuery struct {
	Page, PageSize                                  int
	StartTime, EndTime                              time.Time
	Timezone, SortOrder, Direction, Reason, RefType string
	RefID                                           int64
}

type API interface {
	Login(context.Context) error
	Refresh(context.Context) error
	ListUsage(context.Context, UsageQuery) (Page[UsageRecord], error)
	ListLedger(context.Context, LedgerQuery) (Page[LedgerEntry], error)
}
