package syncer

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Scheduler struct {
	service  *Service
	accounts []string
	interval time.Duration
	mu       sync.Mutex
	running  map[string]bool
}

func NewScheduler(service *Service, accounts []string, interval time.Duration) *Scheduler {
	return &Scheduler{service: service, accounts: accounts, interval: interval, running: map[string]bool{}}
}
func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, id := range s.accounts {
				id := id
				go s.Trigger(ctx, id)
			}
		}
	}
}
func (s *Scheduler) Trigger(ctx context.Context, id string) error {
	s.mu.Lock()
	if s.running[id] {
		s.mu.Unlock()
		return fmt.Errorf("sync already running for %s", id)
	}
	s.running[id] = true
	s.mu.Unlock()
	defer func() { s.mu.Lock(); delete(s.running, id); s.mu.Unlock() }()
	return s.service.SyncAccount(ctx, id)
}
