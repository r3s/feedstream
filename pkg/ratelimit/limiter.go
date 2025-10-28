package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	attempts map[string][]time.Time
	mu       sync.RWMutex
}

func NewLimiter() *Limiter {
	l := &Limiter{
		attempts: make(map[string][]time.Time),
	}
	go l.cleanup()
	return l
}

func (l *Limiter) Allow(key string, maxAttempts int, window time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-window)

	attempts, exists := l.attempts[key]
	if !exists {
		l.attempts[key] = []time.Time{now}
		return true
	}

	var validAttempts []time.Time
	for _, timestamp := range attempts {
		if timestamp.After(cutoff) {
			validAttempts = append(validAttempts, timestamp)
		}
	}

	if len(validAttempts) >= maxAttempts {
		l.attempts[key] = validAttempts
		return false
	}

	validAttempts = append(validAttempts, now)
	l.attempts[key] = validAttempts
	return true
}

func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}

func (l *Limiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for key, attempts := range l.attempts {
			var validAttempts []time.Time
			for _, timestamp := range attempts {
				if now.Sub(timestamp) < 24*time.Hour {
					validAttempts = append(validAttempts, timestamp)
				}
			}
			if len(validAttempts) == 0 {
				delete(l.attempts, key)
			} else {
				l.attempts[key] = validAttempts
			}
		}
		l.mu.Unlock()
	}
}