package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/analytics"
)

func (s *server) overview(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	start, err1 := time.Parse(time.RFC3339, q.Get("start"))
	end, err2 := time.Parse(time.RFC3339, q.Get("end"))
	if err1 != nil || err2 != nil || !end.After(start) {
		writeError(w, 400, "invalid_range", "valid start and end are required")
		return
	}
	result, err := s.deps.Analytics.Overview(r.Context(), analytics.Filters{AccountID: q.Get("account_id"), Start: start, End: end, Granularity: q.Get("granularity")})
	if err != nil {
		writeError(w, 400, "invalid_query", err.Error())
		return
	}
	writeJSON(w, 200, result)
}

func parseFilters(r *http.Request) (analytics.Filters, error) {
	q := r.URL.Query()
	start, err := time.Parse(time.RFC3339, q.Get("start"))
	if err != nil {
		return analytics.Filters{}, err
	}
	end, err := time.Parse(time.RFC3339, q.Get("end"))
	if err != nil || !end.After(start) {
		return analytics.Filters{}, err
	}
	apiKey, _ := strconv.ParseInt(q.Get("api_key_id"), 10, 64)
	refID, _ := strconv.ParseInt(q.Get("ref_id"), 10, 64)
	return analytics.Filters{AccountID: q.Get("account_id"), Start: start, End: end, Granularity: q.Get("granularity"), Reason: q.Get("reason"), Category: q.Get("category"), Direction: q.Get("direction"), Remark: q.Get("remark"), RefType: q.Get("ref_type"), RefID: refID, Model: q.Get("model"), Endpoint: q.Get("endpoint"), APIKeyID: apiKey}, nil
}
func (s *server) usageRecords(w http.ResponseWriter, r *http.Request) {
	f, e := parseFilters(r)
	if e != nil {
		writeError(w, 400, "invalid_range", "valid start and end are required")
		return
	}
	rows, e := s.deps.Analytics.UsageRecords(r.Context(), f)
	if e != nil {
		writeError(w, 500, "query_failed", "query failed")
		return
	}
	writeJSON(w, 200, rows)
}
func (s *server) ledgerEntries(w http.ResponseWriter, r *http.Request) {
	f, e := parseFilters(r)
	if e != nil {
		writeError(w, 400, "invalid_range", "valid start and end are required")
		return
	}
	rows, e := s.deps.Analytics.LedgerEntries(r.Context(), f)
	if e != nil {
		writeError(w, 500, "query_failed", "query failed")
		return
	}
	writeJSON(w, 200, rows)
}
func (s *server) accountStatuses(w http.ResponseWriter, r *http.Request) {
	rows, e := s.deps.Analytics.AccountStatuses(r.Context())
	if e != nil {
		writeError(w, 500, "query_failed", "query failed")
		return
	}
	writeJSON(w, 200, rows)
}
func (s *server) triggerSync(w http.ResponseWriter, r *http.Request) {
	if s.deps.TriggerSync == nil {
		writeError(w, 503, "sync_unavailable", "sync unavailable")
		return
	}
	if e := s.deps.TriggerSync(r.Context(), r.PathValue("id")); e != nil {
		writeError(w, 409, "sync_failed", e.Error())
		return
	}
	writeJSON(w, 202, map[string]bool{"started": true})
}
