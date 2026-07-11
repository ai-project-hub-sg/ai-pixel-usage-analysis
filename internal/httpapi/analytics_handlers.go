package httpapi

import (
	"net/http"
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
