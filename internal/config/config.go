package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Server   ServerConfig    `toml:"server"`
	Analysis AnalysisConfig  `toml:"analysis"`
	Auth     AuthConfig      `toml:"auth"`
	Hosts    []HostConfig    `toml:"host"`
	Accounts []AccountConfig `toml:"account"`
}

type ServerConfig struct {
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	PublicURL    string `toml:"public_url"`
	SecureCookie bool   `toml:"secure_cookie"`
}

func (c ServerConfig) Address() string { return net.JoinHostPort(c.Host, fmt.Sprint(c.Port)) }

type AnalysisConfig struct {
	Timezone                   string `toml:"timezone"`
	SyncInterval               string `toml:"sync_interval"`
	SyncOverlap                string `toml:"sync_overlap"`
	PreferredHostProbeInterval string `toml:"preferred_host_probe_interval"`
}

type AuthConfig struct {
	SessionTTL string `toml:"session_ttl"`
}

type HostConfig struct {
	URL      string `toml:"url"`
	Priority int    `toml:"priority"`
}

type AccountConfig struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	EmailEnv    string `toml:"email_env"`
	PasswordEnv string `toml:"password_env"`
	Enabled     bool   `toml:"enabled"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	cfg := Config{Server: ServerConfig{Host: "127.0.0.1", Port: 8080}, Analysis: AnalysisConfig{Timezone: "Asia/Shanghai", SyncInterval: "1m", SyncOverlap: "5m", PreferredHostProbeInterval: "5m"}, Auth: AuthConfig{SessionTTL: "24h"}}
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) OrderedHosts() []HostConfig {
	hosts := append([]HostConfig(nil), c.Hosts...)
	sort.SliceStable(hosts, func(i, j int) bool { return hosts[i].Priority < hosts[j].Priority })
	return hosts
}

func (c *Config) validate() error {
	if c.Server.Host == "" || c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server address")
	}
	if c.Server.PublicURL != "" {
		if u, err := url.ParseRequestURI(c.Server.PublicURL); err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("invalid server public_url")
		}
	}
	if _, err := time.LoadLocation(c.Analysis.Timezone); err != nil {
		return fmt.Errorf("invalid analysis timezone: %w", err)
	}
	for name, raw := range map[string]string{"sync_interval": c.Analysis.SyncInterval, "sync_overlap": c.Analysis.SyncOverlap, "preferred_host_probe_interval": c.Analysis.PreferredHostProbeInterval} {
		if d, err := time.ParseDuration(raw); err != nil || d <= 0 {
			return fmt.Errorf("invalid %s", name)
		}
	}
	if d, err := time.ParseDuration(c.Auth.SessionTTL); err != nil || d != 24*time.Hour {
		return fmt.Errorf("auth session_ttl must be 24h")
	}
	if len(c.Hosts) == 0 {
		return fmt.Errorf("at least one host is required")
	}
	for i := range c.Hosts {
		u, err := url.Parse(c.Hosts[i].URL)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return fmt.Errorf("invalid host URL at index %d", i)
		}
		c.Hosts[i].URL = strings.TrimRight(u.String(), "/") + "/"
		if c.Hosts[i].Priority < 1 || c.Hosts[i].Priority > 99 {
			return fmt.Errorf("host priority must be between 1 and 99")
		}
	}
	seen := map[string]struct{}{}
	if len(c.Accounts) == 0 {
		return fmt.Errorf("at least one account is required")
	}
	for i, account := range c.Accounts {
		if account.ID == "" || account.EmailEnv == "" || account.PasswordEnv == "" {
			return fmt.Errorf("account %d requires id, email_env, and password_env", i)
		}
		if _, exists := seen[account.ID]; exists {
			return fmt.Errorf("duplicate account id %q", account.ID)
		}
		seen[account.ID] = struct{}{}
	}
	return nil
}
