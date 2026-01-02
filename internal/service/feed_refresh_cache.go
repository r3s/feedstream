package service

import (
	"sync"
	"time"
)

type FeedRefreshCache struct {
	mu              sync.RWMutex
	lastRefreshTime map[int]time.Time
	ttl             time.Duration
}

func NewFeedRefreshCache(ttlHours int) *FeedRefreshCache {
	return &FeedRefreshCache{
		lastRefreshTime: make(map[int]time.Time),
		ttl:             time.Duration(ttlHours) * time.Hour,
	}
}

func (c *FeedRefreshCache) ShouldRefresh(userID int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	lastRefresh, exists := c.lastRefreshTime[userID]
	if !exists {
		return true
	}

	return time.Since(lastRefresh) >= c.ttl
}

func (c *FeedRefreshCache) RecordRefresh(userID int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastRefreshTime[userID] = time.Now()
}

func (c *FeedRefreshCache) GetLastRefreshTime(userID int) *time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if lastRefresh, exists := c.lastRefreshTime[userID]; exists {
		return &lastRefresh
	}
	return nil
}
