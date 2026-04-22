package ratelimiter_test
import (
	"sayakaya/pkg/ratelimiter"
	"testing"
	"time"
)
func TestTokenBucket(t *testing.T) {
	tb := ratelimiter.NewTokenBucket(2, 5)
	if !tb.Allow() {
		t.Error("expected first request to be allowed")
	}
	if !tb.Allow() {
		t.Error("expected second request to be allowed")
	}
	if tb.Allow() {
		t.Error("expected third request to be blocked (capacity 2)")
	}
	tb2 := ratelimiter.NewTokenBucket(1, 600)
	if !tb2.Allow() {
		t.Error("expected allow")
	}
	if tb2.Allow() {
		t.Error("expected block")
	}
	time.Sleep(150 * time.Millisecond)
	if !tb2.Allow() {
		t.Error("expected allow after wait")
	}
}
func TestManager(t *testing.T) {
	m := ratelimiter.NewManager(1, 60)
	if !m.Allow("user1") {
		t.Error("user1 first allow")
	}
	if m.Allow("user1") {
		t.Error("user1 second block")
	}
	if !m.Allow("user2") {
		t.Error("user2 first allow")
	}
}
