package ratelimiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   float64
	refillRate float64
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(capacity float64, refillRatePerMinute float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		refillRate: refillRatePerMinute / 60.0,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

func (b *TokenBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastRefill = now

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}
	return false
}

type Manager struct {
	buckets    map[string]*TokenBucket
	capacity   float64
	refillRate float64
	mu         sync.Mutex
}

func NewManager(capacity float64, refillRatePerMinute float64) *Manager {
	return &Manager{
		buckets:    make(map[string]*TokenBucket),
		capacity:   capacity,
		refillRate: refillRatePerMinute,
	}
}

func (m *Manager) Allow(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[key]; !exists {
		m.buckets[key] = NewTokenBucket(m.capacity, m.refillRate)
	}

	return m.buckets[key].Allow()
}
