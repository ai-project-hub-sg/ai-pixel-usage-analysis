package analytics

import (
	"encoding/json"
	"time"
)

type Filters struct {
	AccountID                                    string
	Start                                        time.Time
	End                                          time.Time
	Granularity                                  string
	Reason, Category, Direction, Remark, RefType string
	RefID                                        int64
	Model, Endpoint                              string
	APIKeyID                                     int64
}

type Comparison struct {
	Requests           int64   `json:"requests"`
	ActualCost         float64 `json:"actual_cost"`
	Credit, Debit, Net float64
}
type Bucket struct {
	Start        time.Time   `json:"start"`
	Requests     int64       `json:"requests"`
	InputTokens  int64       `json:"input_tokens"`
	OutputTokens int64       `json:"output_tokens"`
	ActualCost   float64     `json:"actual_cost"`
	Credit       float64     `json:"credit"`
	Debit        float64     `json:"debit"`
	Net          float64     `json:"net"`
	Yesterday    *Comparison `json:"yesterday"`
	LastWeek     *Comparison `json:"last_week"`
}
type Overview struct {
	Buckets []Bucket `json:"buckets"`
}

type UsageRow struct {
	ID                                    int64 `json:"id"`
	AccountID, Model, Endpoint, CreatedAt string
	APIKeyID, InputTokens, OutputTokens   int64
	ActualCost                            string
}
type LedgerRow struct {
	ID                                                                                       int64 `json:"id"`
	AccountID, Direction, Amount, BalanceAfter, Reason, Category, Remark, RefType, CreatedAt string
	RefID                                                                                    int64
	Metadata                                                                                 json.RawMessage
}
type AccountStatus struct {
	ID, Name                           string
	Enabled                            bool
	CurrentHost, LastSyncAt, LastError string
}
