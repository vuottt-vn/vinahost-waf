package ratelimit

import (
	"sync"
	"time"
)

// RateLimiter implements a sliding window rate limiter
type RateLimiter struct {
	clients map[string]*ClientRate
	mu      sync.RWMutex
	maxReq  int           // Maximum requests allowed
	window  time.Duration // Time window
}

// ClientRate tracks request timestamps for a single client
type ClientRate struct {
	requests []time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*ClientRate),
		maxReq:  maxRequests,
		window:  window,
	}
	// Cleanup expired entries periodically
	go rl.cleanup()
	return rl
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	client, exists := rl.clients[ip]
	if !exists {
		client = &ClientRate{
			requests: make([]time.Time, 0, rl.maxReq),
		}
		rl.clients[ip] = client
	}

	// Remove old requests outside the window
	validRequests := make([]time.Time, 0, len(client.requests))
	for _, t := range client.requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}
	client.requests = validRequests

	// Check if under limit
	if len(client.requests) >= rl.maxReq {
		return false
	}

	// Add current request
	client.requests = append(client.requests, now)
	return true
}

// GetRemaining returns the number of remaining requests for an IP
func (rl *RateLimiter) GetRemaining(ip string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	client, exists := rl.clients[ip]
	if !exists {
		return rl.maxReq
	}

	// Count valid requests
	count := 0
	for _, t := range client.requests {
		if t.After(windowStart) {
			count++
		}
	}

	remaining := rl.maxReq - count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetRetryAfter returns the duration until the client can make another request
func (rl *RateLimiter) GetRetryAfter(ip string) time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	client, exists := rl.clients[ip]
	if !exists || len(client.requests) == 0 {
		return 0
	}

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Find the oldest request in the current window
	for _, t := range client.requests {
		if t.After(windowStart) {
			// This request will expire at t + window
			retryAfter := t.Add(rl.window).Sub(now)
			if retryAfter > 0 {
				return retryAfter
			}
		}
	}

	return 0
}

// Reset clears rate limit data for a specific IP
func (rl *RateLimiter) Reset(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.clients, ip)
}

// ResetAll clears all rate limit data
func (rl *RateLimiter) ResetAll() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.clients = make(map[string]*ClientRate)
}

// GetStats returns rate limit statistics
func (rl *RateLimiter) GetStats() map[string]int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := make(map[string]int)
	now := time.Now()
	windowStart := now.Add(-rl.window)

	for ip, client := range rl.clients {
		count := 0
		for _, t := range client.requests {
			if t.After(windowStart) {
				count++
			}
		}
		if count > 0 {
			stats[ip] = count
		}
	}

	return stats
}

// cleanup removes expired client entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-rl.window)

		for ip, client := range rl.clients {
			// Check if client has any recent requests
			hasRecent := false
			for _, t := range client.requests {
				if t.After(windowStart) {
					hasRecent = true
					break
				}
			}
			if !hasRecent {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}
