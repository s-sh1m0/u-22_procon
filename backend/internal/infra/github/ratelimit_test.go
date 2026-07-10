package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimitTransport_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "4999")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10))
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)

	rlt := newRateLimitTransport(http.DefaultTransport)
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := rlt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("got %d, want 200", resp.StatusCode)
	}
}

func TestRateLimitTransport_Retry429(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			return
		}
		w.Header().Set("X-RateLimit-Remaining", "4998")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10))
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)

	rlt := newRateLimitTransport(http.DefaultTransport)
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := rlt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("got %d, want 200", resp.StatusCode)
	}
	if got := calls.Load(); got != 2 {
		t.Errorf("expected 2 calls, got %d", got)
	}
}

func TestRateLimitTransport_Retry403RateLimit(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(1*time.Second).Unix(), 10))
			w.WriteHeader(403)
			return
		}
		w.Header().Set("X-RateLimit-Remaining", "4999")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10))
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)

	rlt := newRateLimitTransport(http.DefaultTransport)
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := rlt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("got %d, want 200", resp.StatusCode)
	}
	if got := calls.Load(); got != 2 {
		t.Errorf("expected 2 calls, got %d", got)
	}
}

func TestRateLimitTransport_Retry5xx(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.WriteHeader(502)
			return
		}
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)

	rlt := newRateLimitTransport(http.DefaultTransport)
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := rlt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("got %d, want 200", resp.StatusCode)
	}
	if got := calls.Load(); got != 2 {
		t.Errorf("expected 2 calls, got %d", got)
	}
}

func TestRateLimitTransport_MaxRetriesExhausted(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(429)
	}))
	t.Cleanup(srv.Close)

	rlt := newRateLimitTransport(http.DefaultTransport)
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := rlt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 429 {
		t.Fatalf("got %d, want 429 after exhausting retries", resp.StatusCode)
	}
	if got := calls.Load(); int(got) != maxRetries+1 {
		t.Errorf("expected %d calls, got %d", maxRetries+1, got)
	}
}

func TestRateLimitTransport_ContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(429)
	}))
	t.Cleanup(srv.Close)

	rlt := newRateLimitTransport(http.DefaultTransport)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL, nil)
	_, err := rlt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error from canceled context")
	}
}

func TestRateLimitTransport_403NotRateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(403)
	}))
	t.Cleanup(srv.Close)

	rlt := newRateLimitTransport(http.DefaultTransport)
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := rlt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatalf("got %d, want 403 (permission denied, no retry)", resp.StatusCode)
	}
}

func TestRetryWait_RetryAfterHeader(t *testing.T) {
	resp := &http.Response{
		StatusCode: 429,
		Header:     http.Header{"Retry-After": {"30"}},
	}
	got := retryWait(resp, 0)
	if got != 30*time.Second {
		t.Errorf("got %s, want 30s", got)
	}
}

func TestRetryWait_5xxExponentialBackoff(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
	}
	for _, tt := range tests {
		resp := &http.Response{StatusCode: 500, Header: http.Header{}}
		got := retryWait(resp, tt.attempt)
		if got != tt.want {
			t.Errorf("attempt %d: got %s, want %s", tt.attempt, got, tt.want)
		}
	}
}

func TestNewClientFactory_CachesClient(t *testing.T) {
	factory := NewClientFactory()
	c1 := factory(context.Background(), "token-a")
	c2 := factory(context.Background(), "token-a")
	c3 := factory(context.Background(), "token-b")

	if c1 != c2 {
		t.Error("same token should return the same client")
	}
	if c1 == c3 {
		t.Error("different tokens should return different clients")
	}
}

func TestNewClientFactory_EmptyToken(t *testing.T) {
	factory := NewClientFactory()
	c1 := factory(context.Background(), "")
	c2 := factory(context.Background(), "")
	if c1 == c2 {
		t.Error("empty token should not be cached")
	}
}
