package secrets

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/config"
)

const (
	dashboardUsernameKey = "DASHBOARD_USERNAME"
	dashboardPasswordKey = "DASHBOARD_PASSWORD"
)

type DashboardCredentials struct {
	Username string
	Password string
}

type AccountCredentials struct {
	Email    string
	Password string
}

func EnsureDashboardCredentials(path string) (DashboardCredentials, bool, error) {
	b, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return DashboardCredentials{}, false, fmt.Errorf("read dotenv: %w", err)
	}
	original := string(b)
	values := parseDotEnv(original)
	username, password := values[dashboardUsernameKey], values[dashboardPasswordKey]
	generated := false
	if username == "" {
		username, err = randomValue(12)
		if err != nil {
			return DashboardCredentials{}, false, err
		}
		username = "admin-" + username
		original = appendEnv(original, dashboardUsernameKey, username)
		generated = true
	}
	if password == "" {
		password, err = randomValue(24)
		if err != nil {
			return DashboardCredentials{}, false, err
		}
		original = appendEnv(original, dashboardPasswordKey, password)
		generated = true
	}
	if generated {
		if err := os.MkdirAll(directory(path), 0o700); err != nil {
			return DashboardCredentials{}, false, err
		}
		if err := atomicWrite(path, []byte(original)); err != nil {
			return DashboardCredentials{}, false, err
		}
	} else if err := os.Chmod(path, 0o600); err != nil {
		return DashboardCredentials{}, false, fmt.Errorf("secure dotenv permissions: %w", err)
	}
	return DashboardCredentials{Username: username, Password: password}, generated, nil
}

func LoadAccountCredentials(path string, accounts []config.AccountConfig) (map[string]AccountCredentials, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read dotenv: %w", err)
	}
	values := parseDotEnv(string(b))
	result := make(map[string]AccountCredentials, len(accounts))
	for _, account := range accounts {
		email, password := values[account.EmailEnv], values[account.PasswordEnv]
		if email == "" || password == "" {
			return nil, fmt.Errorf("missing credentials for account %q", account.ID)
		}
		result[account.ID] = AccountCredentials{Email: email, Password: password}
	}
	return result, nil
}

func randomValue(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random credential: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func appendEnv(content, key, value string) string {
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + key + "=" + value + "\n"
}
