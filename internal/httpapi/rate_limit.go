package httpapi

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type loginAttempt struct {
	first time.Time
	count int
}

type loginLimiter struct {
	mu       sync.Mutex
	attempts map[string]loginAttempt
	maximum  int
	window   time.Duration
}

func newLoginLimiter(maximum int, window time.Duration) *loginLimiter {
	return &loginLimiter{attempts: make(map[string]loginAttempt), maximum: maximum, window: window}
}

func (limiter *loginLimiter) blocked(key string, now time.Time) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	attempt, exists := limiter.attempts[key]
	if exists && now.Sub(attempt.first) >= limiter.window {
		delete(limiter.attempts, key)
		return false
	}
	return exists && attempt.count >= limiter.maximum
}

func (limiter *loginLimiter) fail(key string, now time.Time) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	if len(limiter.attempts) >= 4096 {
		for candidate, attempt := range limiter.attempts {
			if now.Sub(attempt.first) >= limiter.window {
				delete(limiter.attempts, candidate)
			}
		}
	}
	attempt, exists := limiter.attempts[key]
	if !exists || now.Sub(attempt.first) >= limiter.window {
		limiter.attempts[key] = loginAttempt{first: now, count: 1}
		return
	}
	attempt.count++
	limiter.attempts[key] = attempt
}

func (limiter *loginLimiter) reset(key string) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	delete(limiter.attempts, key)
}

func loginLimitKey(request *http.Request, username string) string {
	return clientIP(request) + "\x00" + strings.ToLower(strings.TrimSpace(username))
}

func clientIP(request *http.Request) string {
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		host = request.RemoteAddr
	}
	peer := net.ParseIP(host)
	if peer != nil && peer.IsLoopback() {
		forwarded := strings.TrimSpace(strings.Split(request.Header.Get("X-Forwarded-For"), ",")[0])
		if parsed := net.ParseIP(forwarded); parsed != nil {
			return parsed.String()
		}
	}
	if peer != nil {
		return peer.String()
	}
	return host
}
