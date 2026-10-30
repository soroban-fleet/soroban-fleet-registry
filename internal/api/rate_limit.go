package api

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimitConfig configures per-client request budgets. A category is disabled
// when its RequestsPerSecond is zero or negative.
type RateLimitConfig struct {
	// RequestsPerSecond and Burst apply to ordinary read endpoints.
	RequestsPerSecond float64
	Burst             float64

	// HeavyRequestsPerSecond and HeavyBurst apply to endpoints that resolve
	// live ledger state through Stellar RPC.
	HeavyRequestsPerSecond float64
	HeavyBurst             float64

	// TrustProxy keys clients by the first X-Forwarded-For hop instead of the
	// peer address. Enable only when every request arrives through a trusted
	// reverse proxy, otherwise clients can forge the header to escape limits.
	TrustProxy bool
}

const (
	// maxTrackedClients bounds limiter memory against address-spoofing floods.
	maxTrackedClients = 50000

	// bucketIdleTTL is how long an unused client bucket is retained.
	bucketIdleTTL = 10 * time.Minute

	// sharedOverflowKey is the fallback bucket used once the client cap is hit.
	sharedOverflowKey = "__overflow__"
)

// rateLimiter holds one token bucket per client for a single budget.
type rateLimiter struct {
	rps   float64
	burst float64

	mu        sync.Mutex
	clients   map[string]*tokenBucket
	lastSweep time.Time
}

func newRateLimiter(rps, burst float64) *rateLimiter {
	if burst < 1 {
		burst = 1
	}
	return &rateLimiter{
		rps:     rps,
		burst:   burst,
		clients: make(map[string]*tokenBucket),
	}
}

// tokenBucket accumulates request credits at a constant refill rate.
type tokenBucket struct {
	tokens    float64
	updatedAt time.Time
}

// allow consumes one credit, or reports how long the client must wait.
func (b *tokenBucket) allow(now time.Time, rps, burst float64) (bool, time.Duration) {
	elapsed := now.Sub(b.updatedAt).Seconds()
	if elapsed > 0 {
		b.tokens = math.Min(burst, b.tokens+elapsed*rps)
	}
	b.updatedAt = now

	if b.tokens >= 1 {
		b.tokens--
		return true, 0
	}
	return false, time.Duration(math.Ceil((1 - b.tokens) / rps * float64(time.Second)))
}

// allow grants a request for the given client, returning false and a retry
// delay once the client is over budget.
func (l *rateLimiter) allow(client string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastSweep) > bucketIdleTTL/2 {
		l.sweep(now)
	}

	bucket, ok := l.clients[client]
	if !ok {
		if len(l.clients) >= maxTrackedClients {
			// New addresses can no longer be tracked; share one bucket so the
			// limiter degrades to throttling instead of exhausting memory.
			client = sharedOverflowKey
			bucket, ok = l.clients[client]
		}
		if !ok {
			bucket = &tokenBucket{tokens: l.burst, updatedAt: now}
			l.clients[client] = bucket
		}
	}

	return bucket.allow(now, l.rps, l.burst)
}

func (l *rateLimiter) sweep(now time.Time) {
	for client, bucket := range l.clients {
		if now.Sub(bucket.updatedAt) > bucketIdleTTL && client != sharedOverflowKey {
			delete(l.clients, client)
		}
	}
	l.lastSweep = now
}

// routeLimiters assigns each request to a budget category.
type routeLimiters struct {
	config  RateLimitConfig
	base    *rateLimiter
	heavy   *rateLimiter
	nowFunc func() time.Time
}

func newRouteLimiters(cfg RateLimitConfig, nowFunc func() time.Time) *routeLimiters {
	if nowFunc == nil {
		nowFunc = time.Now
	}
	rl := &routeLimiters{config: cfg, nowFunc: nowFunc}
	if cfg.RequestsPerSecond > 0 {
		rl.base = newRateLimiter(cfg.RequestsPerSecond, cfg.Burst)
	}
	if cfg.HeavyRequestsPerSecond > 0 {
		rl.heavy = newRateLimiter(cfg.HeavyRequestsPerSecond, cfg.HeavyBurst)
	}
	return rl
}

func (rl *routeLimiters) enabled() bool {
	return rl != nil && (rl.base != nil || rl.heavy != nil)
}

// protect applies the budget for a route category to a handler.
func (rl *routeLimiters) protect(heavy bool, next http.HandlerFunc) http.HandlerFunc {
	if !rl.enabled() {
		return next
	}
	limiter := rl.base
	if heavy {
		limiter = rl.heavy
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if limiter == nil {
			next.ServeHTTP(w, r)
			return
		}
		ok, retryAfter := limiter.allow(clientKey(r, rl.config.TrustProxy), rl.nowFunc())
		if !ok {
			w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
			writeError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Request rate limit exceeded, retry later")
			return
		}
		next.ServeHTTP(w, r)
	}
}

// clientKey derives the address a request is attributed to.
func clientKey(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			first, _, _ := strings.Cut(xff, ",")
			if ip := net.ParseIP(strings.TrimSpace(first)); ip != nil {
				return ip.String()
			}
		}
		if ip := net.ParseIP(strings.TrimSpace(r.Header.Get("X-Real-IP"))); ip != nil {
			return ip.String()
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
