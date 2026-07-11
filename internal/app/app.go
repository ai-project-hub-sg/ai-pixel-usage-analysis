package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/analytics"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/auth"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/config"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/httpapi"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/secrets"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/store"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/syncer"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/upstream"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/webui"
)

const Name = "ai-pixel-usage-analysis"

type Options struct{ ConfigPath, EnvPath, DatabasePath string }
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

func Run(ctx context.Context, opts Options) error {
	if opts.ConfigPath == "" {
		opts.ConfigPath = "config.toml"
	}
	if opts.EnvPath == "" {
		opts.EnvPath = ".env"
	}
	if opts.DatabasePath == "" {
		opts.DatabasePath = filepath.Join("data", "analysis.db")
	}
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
	}
	dashboard, generated, err := secrets.EnsureDashboardCredentials(opts.EnvPath)
	if err != nil {
		return fmt.Errorf("dashboard credentials: %w", err)
	}
	if generated {
		fmt.Fprintf(os.Stderr, "dashboard credentials generated in %s\n", opts.EnvPath)
	}
	accountCreds, err := secrets.LoadAccountCredentials(opts.EnvPath, cfg.Accounts)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(opts.DatabasePath), 0o700); err != nil {
		return err
	}
	db, err := store.Open(opts.DatabasePath)
	if err != nil {
		return err
	}
	defer db.Close()
	clock := systemClock{}
	authService := auth.NewService(db, clock)
	if _, err = authService.SyncDashboardUser(ctx, dashboard); err != nil {
		return err
	}
	repo := syncer.NewRepository(db)
	clients := map[string]upstream.API{}
	ids := []string{}
	for _, account := range cfg.Accounts {
		if !account.Enabled {
			continue
		}
		if err = repo.UpsertAccount(ctx, account.ID, account.Name, true); err != nil {
			return err
		}
		credential := accountCreds[account.ID]
		endpoints := []upstream.Endpoint{}
		for _, host := range cfg.OrderedHosts() {
			endpoints = append(endpoints, upstream.Endpoint{URL: host.URL, API: upstream.NewClient(host.URL, credential.Email, credential.Password, nil)})
		}
		clients[account.ID] = upstream.NewFailover(endpoints)
		ids = append(ids, account.ID)
	}
	location, _ := time.LoadLocation(cfg.Analysis.Timezone)
	overlap, _ := time.ParseDuration(cfg.Analysis.SyncOverlap)
	interval, _ := time.ParseDuration(cfg.Analysis.SyncInterval)
	syncService := syncer.NewService(repo, clients, location, clock, overlap)
	scheduler := syncer.NewScheduler(syncService, ids, interval)
	router := httpapi.NewRouter(httpapi.Dependencies{Auth: authService, Analytics: analytics.NewService(db, location), Clock: clock, PublicURL: cfg.Server.PublicURL, SecureCookie: cfg.Server.SecureCookie, Static: webui.Handler()})
	server := &http.Server{Addr: cfg.Server.Address(), Handler: router, ReadHeaderTimeout: 10 * time.Second}
	go scheduler.Start(ctx)
	for _, id := range ids {
		id := id
		go func() { _ = scheduler.Trigger(ctx, id) }()
	}
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err = <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
