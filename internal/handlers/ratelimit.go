package handlers

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// rateLimiter is a fixed-window, in-memory rate limiter keyed by an arbitrary string.
// Each key gets limit allowances per window duration. Stale windows are cleaned up
// every 5 minutes so memory does not grow unboundedly.
type rateLimiter struct {
	mu      sync.Mutex
	windows map[string]*rateWindow
	limit   int
	window  time.Duration
}

type rateWindow struct {
	count int
	reset time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		windows: make(map[string]*rateWindow),
		limit:   limit,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

// allow returns true if the key is within its rate limit, false if it should be rejected.
func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	w, ok := rl.windows[key]
	if !ok || now.After(w.reset) {
		rl.windows[key] = &rateWindow{count: 1, reset: now.Add(rl.window)}
		return true
	}
	if w.count >= rl.limit {
		return false
	}
	w.count++
	return true
}

// cleanup removes expired windows periodically to prevent unbounded memory growth.
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for k, w := range rl.windows {
			if now.After(w.reset) {
				delete(rl.windows, k)
			}
		}
		rl.mu.Unlock()
	}
}

// clientIP extracts the real client IP from the request, respecting Cloudflare headers.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		if i := strings.Index(ip, ","); i != -1 {
			return strings.TrimSpace(ip[:i])
		}
		return strings.TrimSpace(ip)
	}
	// RemoteAddr is "host:port"
	if i := strings.LastIndex(r.RemoteAddr, ":"); i != -1 {
		return r.RemoteAddr[:i]
	}
	return r.RemoteAddr
}
