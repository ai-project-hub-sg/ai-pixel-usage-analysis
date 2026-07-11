package analytics

import "time"

type Filters struct {
	AccountID                                    string
	Start                                        time.Time
	End                                          time.Time
	Granularity                                  string
	Reason, Category, Direction, Remark, RefType string
	RefID                                        int64
}

type Comparison struct {
	Requests           int64   `json:"requests"`
	ActualCost         float64 `json:"actual_cost"`
	Credit, Debit, Net float64
}
type Bucket struct {
	Start                     time.Time `json:"start"`
	Requests                  int64     `json:"requests"`
	InputTokens, OutputTokens int64
	ActualCost                float64 `json:"actual_cost"`
	Credit, Debit, Net        float64
	Yesterday                 *Comparison `json:"yesterday"`
	LastWeek                  *Comparison `json:"last_week"`
}
type Overview struct {
	Buckets []Bucket `json:"buckets"`
}
