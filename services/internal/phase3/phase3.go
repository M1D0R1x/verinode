// Package phase3 implements the Phase 3 "scale & integrations" logic:
// programmatic API keys, bounded automated credit adjustment, and automated
// replacement/substitution routing within the cure window. The core decisions are
// pure functions so they are unit-testable without a database; the HTTP layer wires
// them to persistence (or runs them in-memory in standalone/demo mode).
package phase3

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// -----------------------------------------------------------------------------
// API keys
// -----------------------------------------------------------------------------

// APIKey is a scoped programmatic credential for ERP/partner integration.
type APIKey struct {
	ID           string    `json:"id"`
	Prefix       string    `json:"key_prefix"`
	Scopes       []string  `json:"scopes"`
	RateLimit    int       `json:"rate_limit_per_min"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	Secret       string    `json:"secret,omitempty"` // returned ONCE at creation
	hash         string
}

// GenerateAPIKey mints a new key. The full secret is returned only once; only its
// SHA-256 hash and prefix are persisted.
func GenerateAPIKey(scopes []string, rateLimit int) (*APIKey, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	secret := "vn_live_" + hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(secret))
	if rateLimit <= 0 {
		rateLimit = 120
	}
	return &APIKey{
		Prefix:    secret[:16],
		Scopes:    scopes,
		RateLimit: rateLimit,
		Status:    "active",
		CreatedAt: time.Now().UTC(),
		Secret:    secret,
		hash:      hex.EncodeToString(sum[:]),
	}, nil
}

// Hash returns the persistable SHA-256 hash of the secret.
func (k *APIKey) Hash() string { return k.hash }

// HashSecret hashes an incoming secret for comparison against a stored hash.
func HashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// -----------------------------------------------------------------------------
// Rate limiter (token bucket, per-key, in-memory)
// -----------------------------------------------------------------------------

type bucket struct {
	tokens   float64
	last     time.Time
	capacity float64
	refill   float64 // tokens per second
}

// RateLimiter is a simple per-key token-bucket limiter suitable for a single node.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{buckets: make(map[string]*bucket)}
}

// Allow reports whether a request for key with the given per-minute limit is allowed.
func (rl *RateLimiter) Allow(key string, perMinute int) bool {
	if perMinute <= 0 {
		perMinute = 120
	}
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok {
		b = &bucket{tokens: float64(perMinute), last: now, capacity: float64(perMinute), refill: float64(perMinute) / 60.0}
		rl.buckets[key] = b
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens = min(b.capacity, b.tokens+elapsed*b.refill)
	b.last = now
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// -----------------------------------------------------------------------------
// Idempotency cache (in-memory, TTL) for integration POSTs
// -----------------------------------------------------------------------------

type idemEntry struct {
	body    []byte
	status  int
	created time.Time
}

// IdempotencyCache stores responses keyed by Idempotency-Key for a TTL window.
type IdempotencyCache struct {
	mu  sync.Mutex
	m   map[string]idemEntry
	ttl time.Duration
}

func NewIdempotencyCache(ttl time.Duration) *IdempotencyCache {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &IdempotencyCache{m: make(map[string]idemEntry), ttl: ttl}
}

// Get returns a cached response for key if present and unexpired.
func (c *IdempotencyCache) Get(key string) ([]byte, int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[key]
	if !ok || time.Since(e.created) > c.ttl {
		return nil, 0, false
	}
	return e.body, e.status, true
}

// Put stores a response for key.
func (c *IdempotencyCache) Put(key string, status int, body []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = idemEntry{body: body, status: status, created: time.Now()}
}

// -----------------------------------------------------------------------------
// Automated credit-limit adjustment (bounded, logged)
// -----------------------------------------------------------------------------

// CreditHistory summarizes a buyer's recent payment behaviour.
type CreditHistory struct {
	OnTimeSettlements int   `json:"on_time_settlements"`
	LateSettlements   int   `json:"late_settlements"`
	Defaults          int   `json:"defaults"`
	CurrentLimitCents int64 `json:"current_limit_cents"`
}

// CreditDecision is the bounded, explainable output of the auto-adjuster.
type CreditDecision struct {
	NewLimitCents  int64  `json:"new_limit_cents"`
	DeltaCents     int64  `json:"delta_cents"`
	Reason         string `json:"reason"`
	RequiresReview bool   `json:"requires_review"`
}

// Bounds for a single automated adjustment. Anything larger requires human review.
const (
	maxAutoIncreasePct = 0.20 // at most +20% per adjustment
	maxAutoDecreasePct = 0.50 // at most -50% per adjustment automatically
)

// AdjustCredit computes a bounded credit-limit change from payment history.
// Any adjustment that would exceed the automated bounds is flagged for review
// and returns a zero delta (no silent large change).
func AdjustCredit(h CreditHistory) CreditDecision {
	if h.CurrentLimitCents <= 0 {
		return CreditDecision{Reason: "no baseline limit; manual approval required", RequiresReview: true}
	}

	// A default triggers an immediate bounded decrease and mandatory review.
	if h.Defaults > 0 {
		delta := -int64(float64(h.CurrentLimitCents) * maxAutoDecreasePct)
		return CreditDecision{
			NewLimitCents:  h.CurrentLimitCents + delta,
			DeltaCents:     delta,
			Reason:         fmt.Sprintf("%d default(s) on record: bounded -%.0f%% and escalated", h.Defaults, maxAutoDecreasePct*100),
			RequiresReview: true,
		}
	}

	total := h.OnTimeSettlements + h.LateSettlements
	if total < 5 {
		return CreditDecision{NewLimitCents: h.CurrentLimitCents, Reason: "insufficient settlement history (<5) for automated change"}
	}

	onTimeRate := float64(h.OnTimeSettlements) / float64(total)
	switch {
	case onTimeRate >= 0.95:
		delta := int64(float64(h.CurrentLimitCents) * maxAutoIncreasePct)
		return CreditDecision{
			NewLimitCents: h.CurrentLimitCents + delta,
			DeltaCents:    delta,
			Reason:        fmt.Sprintf("on-time rate %.0f%% over %d settlements: +%.0f%%", onTimeRate*100, total, maxAutoIncreasePct*100),
		}
	case onTimeRate < 0.75:
		delta := -int64(float64(h.CurrentLimitCents) * 0.10)
		return CreditDecision{
			NewLimitCents: h.CurrentLimitCents + delta,
			DeltaCents:    delta,
			Reason:        fmt.Sprintf("on-time rate %.0f%% below 75%%: -10%%", onTimeRate*100),
		}
	default:
		return CreditDecision{NewLimitCents: h.CurrentLimitCents, Reason: fmt.Sprintf("on-time rate %.0f%% within neutral band; no change", onTimeRate*100)}
	}
}
