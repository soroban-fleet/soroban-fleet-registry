package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func newTestServer(cfg RateLimitConfig, now *time.Time) http.Handler {
	server := NewServer(":8080", nil, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	server.nowFunc = func() time.Time { return *now }
	server.SetRateLimit(cfg)
	return server.Handler()
}

func doRequest(handler http.Handler, path, remoteAddr string) *http.Response {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = remoteAddr
	handler.ServeHTTP(rec, req)
	return rec.Result()
}

func TestTokenBucketRefillAndReject(t *testing.T) {
	start := time.Unix(1700000000, 0)
	bucket := &tokenBucket{tokens: 2, updatedAt: start}

	for i := 0; i < 2; i++ {
		ok, wait := bucket.allow(start, 1, 2)
		if !ok {
			t.Fatalf("request %d should be allowed by burst", i)
		}
		if wait != 0 {
			t.Fatalf("expected no wait for allowed request, got %v", wait)
		}
	}

	ok, wait := bucket.allow(start, 1, 2)
	if ok {
		t.Fatalf("expected third immediate request to be rejected")
	}
	if wait <= 0 || wait > time.Second {
		t.Errorf("expected retry-after in (0, 1s], got %v", wait)
	}

	// Refill of a full credit admits the next request.
	if ok, _ := bucket.allow(start.Add(1500*time.Millisecond), 1, 2); !ok {
		t.Fatalf("expected request to be allowed after refill")
	}
}

func TestTokenBucketDoesNotExceedBurst(t *testing.T) {
	start := time.Unix(1700000000, 0)
	bucket := &tokenBucket{tokens: 1, updatedAt: start}

	// Two hours of idle refill must still cap at the burst size of 3.
	if ok, _ := bucket.allow(start.Add(2*time.Hour), 10, 3); !ok {
		t.Fatalf("expected request to be allowed")
	}
	if bucket.tokens != 2 {
		t.Errorf("expected tokens capped at burst-1=2, got %v", bucket.tokens)
	}
}

func TestRateLimiterEvictsIdleClients(t *testing.T) {
	start := time.Unix(1700000000, 0)
	limiter := newRateLimiter(10, 10)
	limiter.lastSweep = start

	if ok, _ := limiter.allow("203.0.113.1", start); !ok {
		t.Fatalf("expected first request to be allowed")
	}
	if len(limiter.clients) != 1 {
		t.Fatalf("expected one tracked client, got %d", len(limiter.clients))
	}

	// A request after the idle TTL triggers a sweep that drops the stale bucket.
	if ok, _ := limiter.allow("198.51.100.7", start.Add(bucketIdleTTL+time.Minute)); !ok {
		t.Fatalf("expected request to be allowed")
	}
	if len(limiter.clients) != 1 {
		t.Errorf("expected idle client bucket to be evicted, found %d", len(limiter.clients))
	}
}

func TestRateLimiterFallsBackToSharedBucketAtCapacity(t *testing.T) {
	start := time.Unix(1700000000, 0)
	limiter := newRateLimiter(1, 1)
	limiter.lastSweep = start

	for i := 0; i < maxTrackedClients; i++ {
		limiter.clients[strconv.Itoa(i)] = &tokenBucket{tokens: 0, updatedAt: start}
	}

	// Unknown clients beyond the cap share one bucket: the first passes, the next is throttled.
	if ok, _ := limiter.allow("untracked-a", start); !ok {
		t.Fatalf("expected first overflow request to be allowed")
	}
	if ok, _ := limiter.allow("untracked-b", start); ok {
		t.Fatalf("expected second overflow request from a different address to share the throttle")
	}
	if len(limiter.clients) != maxTrackedClients+1 {
		t.Errorf("expected client map to stay bounded, got %d entries", len(limiter.clients))
	}
}

func TestClientKey(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		trustProxy bool
		want       string
	}{
		{
			name:       "peer address without proxy trust",
			remoteAddr: "203.0.113.9:55555",
			headers:    map[string]string{"X-Forwarded-For": "198.51.100.1"},
			trustProxy: false,
			want:       "203.0.113.9",
		},
		{
			name:       "first forwarded hop when proxy is trusted",
			remoteAddr: "10.0.0.1:55555",
			headers:    map[string]string{"X-Forwarded-For": "198.51.100.1, 203.0.113.9"},
			trustProxy: true,
			want:       "198.51.100.1",
		},
		{
			name:       "real ip fallback when proxy is trusted",
			remoteAddr: "10.0.0.1:55555",
			headers:    map[string]string{"X-Real-IP": "192.0.2.8"},
			trustProxy: true,
			want:       "192.0.2.8",
		},
		{
			name:       "non address forwarded value is ignored",
			remoteAddr: "203.0.113.9:55555",
			headers:    map[string]string{"X-Forwarded-For": "not-an-ip"},
			trustProxy: true,
			want:       "203.0.113.9",
		},
		{
			name:       "address without port",
			remoteAddr: "203.0.113.9",
			trustProxy: false,
			want:       "203.0.113.9",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/v1/fleets", nil)
			r.RemoteAddr = tc.remoteAddr
			for k, v := range tc.headers {
				r.Header.Set(k, v)
			}
			if got := clientKey(r, tc.trustProxy); got != tc.want {
				t.Errorf("clientKey() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRateLimitMiddlewareThrottlesRPCIntensiveEndpoints(t *testing.T) {
	now := time.Unix(1700000000, 0)
	handler := newTestServer(RateLimitConfig{
		RequestsPerSecond:      10,
		Burst:                  10,
		HeavyRequestsPerSecond: 1,
		HeavyBurst:             1,
	}, &now)

	const client = "203.0.113.10:40000"

	if resp := doRequest(handler, "/v1/indexer/status", client); resp.StatusCode != http.StatusOK {
		t.Fatalf("expected first indexer status request to succeed, got %d", resp.StatusCode)
	}

	resp := doRequest(handler, "/status", client)
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected second RPC-intensive request to be limited, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header on 429 response")
	}

	// Health probes are exempt and must never be throttled.
	for i := 0; i < 5; i++ {
		if resp := doRequest(handler, "/health", client); resp.StatusCode != http.StatusOK {
			t.Fatalf("expected health checks to stay available, got %d", resp.StatusCode)
		}
	}

	// Waiting past the heavy refill interval restores capacity.
	now = now.Add(2 * time.Second)
	if resp := doRequest(handler, "/v1/indexer/status", client); resp.StatusCode != http.StatusOK {
		t.Errorf("expected request to be allowed after refill, got %d", resp.StatusCode)
	}
}

func TestRateLimitMiddlewareSeparatesReadBudget(t *testing.T) {
	now := time.Unix(1700000000, 0)
	handler := newTestServer(RateLimitConfig{
		RequestsPerSecond:      1,
		Burst:                  1,
		HeavyRequestsPerSecond: 1,
		HeavyBurst:             1,
	}, &now)

	const client = "203.0.113.13:40000"

	// Exhaust the read budget; the handler itself has no dependencies wired,
	// so an admitted request answers 500 rather than 429.
	if resp := doRequest(handler, "/v1/fleets", client); resp.StatusCode == http.StatusTooManyRequests {
		t.Fatalf("expected the first read request to be admitted")
	}
	if resp := doRequest(handler, "/v1/fleets", client); resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected the second read request to be limited, got %d", resp.StatusCode)
	}

	// The heavy bucket for this client is still full.
	if resp := doRequest(handler, "/v1/indexer/status", client); resp.StatusCode != http.StatusOK {
		t.Errorf("expected RPC-intensive budget to be independent of the read budget, got %d", resp.StatusCode)
	}
}

func TestRateLimitMiddlewareIsDisabledByZeroRates(t *testing.T) {
	now := time.Unix(1700000000, 0)
	handler := newTestServer(RateLimitConfig{}, &now)

	for i := 0; i < 50; i++ {
		if resp := doRequest(handler, "/v1/indexer/status", "203.0.113.11:40000"); resp.StatusCode == http.StatusTooManyRequests {
			t.Fatalf("expected no throttling when rate limits are disabled")
		}
	}
}

func TestRateLimitedErrorPayload(t *testing.T) {
	now := time.Unix(1700000000, 0)
	handler := newTestServer(RateLimitConfig{HeavyRequestsPerSecond: 1, HeavyBurst: 1}, &now)

	doRequest(handler, "/v1/indexer/status", "203.0.113.12:40000")
	resp := doRequest(handler, "/v1/indexer/status", "203.0.113.12:40000")
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.StatusCode)
	}

	var payload ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode error payload: %v", err)
	}
	if payload.Error.Code != "RATE_LIMITED" {
		t.Errorf("expected error code RATE_LIMITED, got %q", payload.Error.Code)
	}
	if payload.Error.Message == "" {
		t.Errorf("expected a human readable error message")
	}
}

func TestRateLimitIsPerClient(t *testing.T) {
	now := time.Unix(1700000000, 0)
	handler := newTestServer(RateLimitConfig{HeavyRequestsPerSecond: 1, HeavyBurst: 1}, &now)

	if got := doRequest(handler, "/v1/indexer/status", "203.0.113.20:40000").StatusCode; got != http.StatusOK {
		t.Fatalf("expected first client request to succeed, got %d", got)
	}
	if got := doRequest(handler, "/v1/indexer/status", "198.51.100.20:40000").StatusCode; got != http.StatusOK {
		t.Fatalf("expected a different client to keep its own budget, got %d", got)
	}
	if got := doRequest(handler, "/v1/indexer/status", "203.0.113.20:40000").StatusCode; got != http.StatusTooManyRequests {
		t.Fatalf("expected the first client to be limited, got %d", got)
	}
}
