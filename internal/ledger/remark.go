package ledger

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Remark struct {
	Category   string            `json:"category"`
	Text       string            `json:"text"`
	SearchText string            `json:"search_text"`
	Fields     map[string]string `json:"fields"`
}

func Normalize(reason, refType string, refID int64, metadata json.RawMessage) Remark {
	category := categoryFor(reason)
	fields := make(map[string]string)
	var values map[string]any
	if json.Unmarshal(metadata, &values) == nil {
		for key, value := range values {
			switch value := value.(type) {
			case string:
				fields[key] = value
			case float64, bool, json.Number:
				fields[key] = fmt.Sprint(value)
			}
		}
	}
	fields["reason"] = reason
	if refType != "" {
		fields["ref_type"] = refType
	}
	if refID > 0 {
		fields["ref_id"] = fmt.Sprint(refID)
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := []string{reason}
	for _, key := range keys {
		if key != "reason" && fields[key] != "" {
			parts = append(parts, key+"="+fields[key])
		}
	}
	text := strings.Join(parts, " · ")
	return Remark{Category: category, Text: text, SearchText: strings.ToLower(text), Fields: fields}
}

func categoryFor(reason string) string {
	switch reason {
	case "usage_charge":
		return "usage"
	case "redeem_code", "admin_adjustment":
		return "credit"
	case "account_share_income", "invite_share_income", "private_group_commission":
		return "sharing"
	case "account_share_mode_seat_prepay", "account_share_mode_seat_refund", "account_share_mode_seat_waiver_refund", "account_share_mode_income":
		return "subscription"
	default:
		return "other"
	}
}
