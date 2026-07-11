package auth

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/secrets"
	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/store"
)

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

func TestPasswordHashUsesIndependentSalts(t *testing.T) {
	first, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	second, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("password hashes reused a salt")
	}
	if !VerifyPassword(first, "correct horse battery staple") || VerifyPassword(first, "wrong") {
		t.Fatal("password verification mismatch")
	}
}

func TestSessionExpiresAt24Hours(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	clock := &fakeClock{now: time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)}
	svc := NewService(db, clock)
	creds := secrets.DashboardCredentials{Username: "admin", Password: "long-random-password"}
	changed, err := svc.SyncDashboardUser(ctx, creds)
	if err != nil || !changed {
		t.Fatalf("sync changed=%v err=%v", changed, err)
	}
	session, err := svc.Login(ctx, creds.Username, creds.Password)
	if err != nil {
		t.Fatal(err)
	}
	if session.ExpiresAt.Sub(clock.now) != 24*time.Hour {
		t.Fatalf("ttl=%v", session.ExpiresAt.Sub(clock.now))
	}
	if _, err := svc.Authenticate(ctx, session.Token); err != nil {
		t.Fatal(err)
	}
	clock.now = clock.now.Add(24 * time.Hour)
	if _, err := svc.Authenticate(ctx, session.Token); err == nil {
		t.Fatal("session remained valid at expiry")
	}
}

func TestCredentialRotationRevokesSessions(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	clock := &fakeClock{now: time.Now().UTC()}
	svc := NewService(db, clock)
	old := secrets.DashboardCredentials{Username: "admin", Password: "old-password"}
	if _, err := svc.SyncDashboardUser(ctx, old); err != nil {
		t.Fatal(err)
	}
	session, err := svc.Login(ctx, old.Username, old.Password)
	if err != nil {
		t.Fatal(err)
	}
	changed, err := svc.SyncDashboardUser(ctx, secrets.DashboardCredentials{Username: "admin", Password: "new-password"})
	if err != nil || !changed {
		t.Fatalf("rotation changed=%v err=%v", changed, err)
	}
	if _, err := svc.Authenticate(ctx, session.Token); err == nil {
		t.Fatal("old session survived credential rotation")
	}
}
