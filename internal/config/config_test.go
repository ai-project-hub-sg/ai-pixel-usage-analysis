package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validConfig = `
[server]
host = "127.0.0.1"
port = 9090
public_url = "https://usage.example.com"
secure_cookie = true

[analysis]
timezone = "Asia/Shanghai"
sync_interval = "1m"
sync_overlap = "5m"
preferred_host_probe_interval = "5m"

[auth]
session_ttl = "24h"

[[host]]
url = "https://second.example/"
priority = 2
[[host]]
url = "https://first-a.example/"
priority = 1
[[host]]
url = "https://first-b.example/"
priority = 1

[[account]]
id = "primary"
name = "主账户"
email_env = "user1"
password_env = "password1"
enabled = true
`

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadParsesServerAndStableHostPriority(t *testing.T) {
	cfg, err := Load(writeConfig(t, validConfig))
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Server.Address(); got != "127.0.0.1:9090" {
		t.Fatalf("address = %q", got)
	}
	hosts := cfg.OrderedHosts()
	want := []string{"https://first-a.example/", "https://first-b.example/", "https://second.example/"}
	for i := range want {
		if hosts[i].URL != want[i] {
			t.Fatalf("host %d = %q, want %q", i, hosts[i].URL, want[i])
		}
	}
}

func TestLoadRejectsDuplicateAccountIDs(t *testing.T) {
	content := validConfig + `
[[account]]
id = "primary"
name = "重复"
email_env = "user2"
password_env = "password2"
enabled = true
`
	if _, err := Load(writeConfig(t, content)); err == nil {
		t.Fatal("expected duplicate account error")
	}
}

func TestLoadRejectsSessionTTLNot24Hours(t *testing.T) {
	content := strings.Replace(validConfig, `session_ttl = "24h"`, `session_ttl = "12h"`, 1)
	if _, err := Load(writeConfig(t, content)); err == nil {
		t.Fatal("expected session TTL error")
	}
}

func TestLoadRejectsInvalidTimezone(t *testing.T) {
	content := strings.Replace(validConfig, `timezone = "Asia/Shanghai"`, `timezone = "Mars/Olympus"`, 1)
	if _, err := Load(writeConfig(t, content)); err == nil {
		t.Fatal("expected timezone error")
	}
}

func TestLoadRejectsMissingCredentialKey(t *testing.T) {
	content := strings.Replace(validConfig, `password_env = "password1"`, `password_env = ""`, 1)
	if _, err := Load(writeConfig(t, content)); err == nil {
		t.Fatal("expected credential key error")
	}
}

func TestLoadRejectsNoAccounts(t *testing.T) {
	start := strings.Index(validConfig, "[[account]]")
	if _, err := Load(writeConfig(t, validConfig[:start])); err == nil {
		t.Fatal("expected at least one account error")
	}
}
