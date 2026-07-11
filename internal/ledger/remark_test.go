package ledger

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeKnownAndUnknownLedgerRemarks(t *testing.T) {
	tests := []struct {
		name, reason, category, contains, field string
		metadata                                json.RawMessage
	}{
		{"usage", "usage_charge", "usage", "req-7", "request_id", json.RawMessage(`{"request_id":"req-7","api_key_id":12,"account_id":3}`)},
		{"redeem", "redeem_code", "credit", "SPRING", "code", json.RawMessage(`{"code":"SPRING"}`)},
		{"share", "account_share_income", "sharing", "group-4", "group_id", json.RawMessage(`{"group_id":"group-4"}`)},
		{"subscription", "account_share_mode_seat_prepay", "subscription", "2026-07", "period_started", json.RawMessage(`{"period_started":"2026-07-01","period_ended":"2026-07-31","hourly_rate":0.2}`)},
		{"unknown", "future_bonus", "other", "future_bonus", "", json.RawMessage(`{"new_field":"kept"}`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.reason, "usage_log", 9, tt.metadata)
			if got.Category != tt.category || !strings.Contains(got.SearchText, strings.ToLower(tt.contains)) {
				t.Fatalf("remark=%#v", got)
			}
			if tt.field != "" && got.Fields[tt.field] == "" {
				t.Fatalf("missing field %s in %#v", tt.field, got.Fields)
			}
		})
	}
}
