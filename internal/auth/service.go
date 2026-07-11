package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/secrets"
)

var ErrUnauthenticated = errors.New("unauthenticated")

type Clock interface{ Now() time.Time }

type Service struct {
	db    *sql.DB
	clock Clock
}

type Session struct {
	Token     string
	ExpiresAt time.Time
}

type User struct {
	ID       int64
	Username string
}

func NewService(db *sql.DB, clock Clock) *Service {
	return &Service{db: db, clock: clock}
}

func (s *Service) SyncDashboardUser(ctx context.Context, creds secrets.DashboardCredentials) (bool, error) {
	var username, existingHash string
	err := s.db.QueryRowContext(ctx, `SELECT username, password_hash FROM dashboard_users WHERE id=1`).Scan(&username, &existingHash)
	if err == nil && username == creds.Username && VerifyPassword(existingHash, creds.Password) {
		return false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("read dashboard user: %w", err)
	}
	passwordHash, err := HashPassword(creds.Password)
	if err != nil {
		return false, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	now := s.clock.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `INSERT INTO dashboard_users(id,username,password_hash,updated_at) VALUES(1,?,?,?) ON CONFLICT(id) DO UPDATE SET username=excluded.username,password_hash=excluded.password_hash,updated_at=excluded.updated_at`, creds.Username, passwordHash, now); err != nil {
		return false, fmt.Errorf("store dashboard user: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE web_sessions SET revoked_at=? WHERE revoked_at IS NULL`, now); err != nil {
		return false, fmt.Errorf("revoke sessions: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (Session, error) {
	var id int64
	var storedUsername, passwordHash string
	if err := s.db.QueryRowContext(ctx, `SELECT id,username,password_hash FROM dashboard_users WHERE id=1`).Scan(&id, &storedUsername, &passwordHash); err != nil {
		return Session{}, ErrUnauthenticated
	}
	if storedUsername != username || !VerifyPassword(passwordHash, password) {
		return Session{}, ErrUnauthenticated
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return Session{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(token))
	now := s.clock.Now().UTC()
	expires := now.Add(24 * time.Hour)
	if _, err := s.db.ExecContext(ctx, `INSERT INTO web_sessions(token_hash,user_id,created_at,expires_at) VALUES(?,?,?,?)`, hash[:], id, now.Format(time.RFC3339Nano), expires.Format(time.RFC3339Nano)); err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	return Session{Token: token, ExpiresAt: expires}, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (User, error) {
	hash := sha256.Sum256([]byte(token))
	var user User
	var expiresRaw string
	err := s.db.QueryRowContext(ctx, `SELECT u.id,u.username,s.expires_at FROM web_sessions s JOIN dashboard_users u ON u.id=s.user_id WHERE s.token_hash=? AND s.revoked_at IS NULL`, hash[:]).Scan(&user.ID, &user.Username, &expiresRaw)
	if err != nil {
		return User{}, ErrUnauthenticated
	}
	expires, err := time.Parse(time.RFC3339Nano, expiresRaw)
	if err != nil || !s.clock.Now().UTC().Before(expires) {
		return User{}, ErrUnauthenticated
	}
	return user, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	hash := sha256.Sum256([]byte(token))
	_, err := s.db.ExecContext(ctx, `UPDATE web_sessions SET revoked_at=? WHERE token_hash=? AND revoked_at IS NULL`, s.clock.Now().UTC().Format(time.RFC3339Nano), hash[:])
	return err
}
