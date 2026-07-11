package syncer

import (
	"context"
	"fmt"
	"time"

	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/upstream"
)

type Clock interface{ Now() time.Time }
type Service struct {
	repo     *Repository
	clients  map[string]upstream.API
	location *time.Location
	clock    Clock
	overlap  time.Duration
}

func NewService(repo *Repository, clients map[string]upstream.API, location *time.Location, clock Clock, overlap time.Duration) *Service {
	return &Service{repo: repo, clients: clients, location: location, clock: clock, overlap: overlap}
}

func (s *Service) SyncAccount(ctx context.Context, accountID string) (syncErr error) {
	client, ok := s.clients[accountID]
	if !ok {
		return fmt.Errorf("unknown account %q", accountID)
	}
	end := s.clock.Now().UTC()
	defer func() {
		host := ""
		if reporter, ok := client.(interface{ CurrentHost() string }); ok {
			host = reporter.CurrentHost()
		}
		statusErr := s.repo.RecordAccountSync(context.WithoutCancel(ctx), accountID, host, end, syncErr)
		if syncErr == nil && statusErr != nil {
			syncErr = statusErr
		}
	}()
	for _, dataType := range []string{"usage", "ledger"} {
		start, exists, err := s.repo.Cursor(ctx, accountID, dataType)
		if err != nil {
			return err
		}
		if exists {
			start = start.Add(-s.overlap)
		} else {
			local := end.In(s.location)
			start = time.Date(local.Year(), local.Month()-1, 1, 0, 0, 0, 0, s.location).UTC()
		}
		if dataType == "usage" {
			if err = s.syncUsage(ctx, client, accountID, start, end); err != nil {
				return err
			}
		} else {
			if err = s.syncLedger(ctx, client, accountID, start, end); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) syncUsage(ctx context.Context, client upstream.API, accountID string, start, end time.Time) error {
	items := []upstream.UsageRecord{}
	for page := 1; ; page++ {
		result, err := client.ListUsage(ctx, upstream.UsageQuery{Page: page, PageSize: 100, StartTime: start, EndTime: end, Timezone: s.location.String(), SortOrder: "desc"})
		if err != nil {
			return err
		}
		items = append(items, result.Items...)
		if page >= result.Pages || result.Pages == 0 {
			break
		}
	}
	return s.repo.StoreUsage(ctx, accountID, items, end)
}
func (s *Service) syncLedger(ctx context.Context, client upstream.API, accountID string, start, end time.Time) error {
	items := []upstream.LedgerEntry{}
	for page := 1; ; page++ {
		result, err := client.ListLedger(ctx, upstream.LedgerQuery{Page: page, PageSize: 100, StartTime: start, EndTime: end, Timezone: s.location.String(), SortOrder: "desc"})
		if err != nil {
			return err
		}
		items = append(items, result.Items...)
		if page >= result.Pages || result.Pages == 0 {
			break
		}
	}
	return s.repo.StoreLedger(ctx, accountID, items, end)
}
