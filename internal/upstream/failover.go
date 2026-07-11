package upstream

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Endpoint struct {
	URL string
	API API
}

type Failover struct {
	mu            sync.RWMutex
	endpoints     []Endpoint
	current       int
	authenticated bool
}

func NewFailover(endpoints []Endpoint) *Failover {
	return &Failover{endpoints: append([]Endpoint(nil), endpoints...)}
}

func (f *Failover) CurrentHost() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if len(f.endpoints) == 0 {
		return ""
	}
	return f.endpoints[f.current].URL
}

func (f *Failover) Login(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.loginLocked(ctx)
}

func (f *Failover) loginLocked(ctx context.Context) error {
	if len(f.endpoints) == 0 {
		return errors.New("no upstream endpoints")
	}
	f.authenticated = false
	var last error
	for i := range f.endpoints {
		if err := f.endpoints[i].API.Login(ctx); err == nil {
			f.current = i
			f.authenticated = true
			return nil
		} else if IsKind(err, ErrorAuth) || IsKind(err, ErrorClient) {
			return err
		} else {
			last = err
		}
	}
	return fmt.Errorf("all upstream endpoints failed: %w", last)
}

func (f *Failover) Refresh(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	api := f.currentAPI()
	if api == nil {
		return errors.New("no upstream endpoints")
	}
	err := api.Refresh(ctx)
	f.authenticated = err == nil
	return err
}

func (f *Failover) ListUsage(ctx context.Context, q UsageQuery) (Page[UsageRecord], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var zero Page[UsageRecord]
	if !f.authenticated {
		if err := f.loginLocked(ctx); err != nil {
			return zero, err
		}
	}
	for attempts := 0; attempts < len(f.endpoints); attempts++ {
		page, err := f.endpoints[f.current].API.ListUsage(ctx, q)
		if err == nil {
			return page, nil
		}
		if IsKind(err, ErrorAuth) {
			if recoveryErr := f.recoverAuthenticationLocked(ctx); recoveryErr != nil {
				return zero, recoveryErr
			}
			page, err = f.endpoints[f.current].API.ListUsage(ctx, q)
			if err == nil {
				return page, nil
			}
		}
		if !switchable(err) {
			return zero, err
		}
		f.current = (f.current + 1) % len(f.endpoints)
		f.authenticated = false
		if loginErr := f.endpoints[f.current].API.Login(ctx); loginErr != nil {
			if !switchable(loginErr) {
				return zero, loginErr
			}
			continue
		}
		f.authenticated = true
	}
	return zero, errors.New("all upstream endpoints failed")
}

func (f *Failover) ListLedger(ctx context.Context, q LedgerQuery) (Page[LedgerEntry], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var zero Page[LedgerEntry]
	if !f.authenticated {
		if err := f.loginLocked(ctx); err != nil {
			return zero, err
		}
	}
	for attempts := 0; attempts < len(f.endpoints); attempts++ {
		page, err := f.endpoints[f.current].API.ListLedger(ctx, q)
		if err == nil {
			return page, nil
		}
		if IsKind(err, ErrorAuth) {
			if recoveryErr := f.recoverAuthenticationLocked(ctx); recoveryErr != nil {
				return zero, recoveryErr
			}
			page, err = f.endpoints[f.current].API.ListLedger(ctx, q)
			if err == nil {
				return page, nil
			}
		}
		if !switchable(err) {
			return zero, err
		}
		f.current = (f.current + 1) % len(f.endpoints)
		f.authenticated = false
		if loginErr := f.endpoints[f.current].API.Login(ctx); loginErr != nil {
			if !switchable(loginErr) {
				return zero, loginErr
			}
			continue
		}
		f.authenticated = true
	}
	return zero, errors.New("all upstream endpoints failed")
}

func (f *Failover) ProbePreferred(ctx context.Context) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := 0; i < f.current; i++ {
		if f.endpoints[i].API.Login(ctx) == nil {
			f.current = i
			f.authenticated = true
			return true
		}
	}
	return false
}

func (f *Failover) recoverAuthenticationLocked(ctx context.Context) error {
	api := f.currentAPI()
	if api == nil {
		return errors.New("no upstream endpoints")
	}
	if err := api.Refresh(ctx); err == nil {
		f.authenticated = true
		return nil
	}
	if err := api.Login(ctx); err != nil {
		f.authenticated = false
		return err
	}
	f.authenticated = true
	return nil
}

func (f *Failover) currentAPI() API {
	if len(f.endpoints) == 0 {
		return nil
	}
	return f.endpoints[f.current].API
}

func switchable(err error) bool {
	return IsKind(err, ErrorTransport) || IsKind(err, ErrorServer) || IsKind(err, ErrorRateLimit)
}
