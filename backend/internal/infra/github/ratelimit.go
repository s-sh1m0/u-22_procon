package github

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const maxRetries = 3

// rateLimitTransport は http.RoundTripper をラップし、GitHub API の
// レート制限(429/403)と一時的なサーバーエラー(5xx)をリトライする。
type rateLimitTransport struct {
	base      http.RoundTripper
	mu        sync.Mutex
	remaining int
	resetAt   time.Time
}

func newRateLimitTransport(base http.RoundTripper) *rateLimitTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &rateLimitTransport{
		base:      base,
		remaining: -1,
	}
}

// RoundTrip はリクエストを送信し、レート制限や 5xx に対してリトライする。
func (t *rateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.waitIfExhausted(req)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			if err := req.Context().Err(); err != nil {
				return nil, err
			}
		}

		r := req
		if attempt > 0 {
			r = req.Clone(req.Context())
		}

		resp, err := t.base.RoundTrip(r)
		if err != nil {
			return nil, err
		}

		t.updateRateLimit(resp)

		if !isRetryableStatus(resp) || attempt == maxRetries {
			return resp, nil
		}

		wait := retryWait(resp, attempt)
		log.Printf("github: %d response, retry in %s (%d/%d)",
			resp.StatusCode, wait, attempt+1, maxRetries)
		_ = resp.Body.Close()

		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
		case <-req.Context().Done():
			timer.Stop()
			return nil, req.Context().Err()
		}
	}
	return nil, req.Context().Err()
}

func (t *rateLimitTransport) waitIfExhausted(req *http.Request) {
	t.mu.Lock()
	remaining := t.remaining
	resetAt := t.resetAt
	t.mu.Unlock()

	if remaining != 0 || resetAt.IsZero() {
		return
	}

	wait := time.Until(resetAt)
	if wait <= 0 {
		return
	}
	const maxWait = 120 * time.Second
	if wait > maxWait {
		wait = maxWait
	}

	log.Printf("github: rate limit exhausted, waiting %s", wait)
	timer := time.NewTimer(wait)
	select {
	case <-timer.C:
	case <-req.Context().Done():
		timer.Stop()
	}
}

func (t *rateLimitTransport) updateRateLimit(resp *http.Response) {
	rem, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
	if err != nil {
		return
	}
	resetUnix, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64)
	if err != nil {
		return
	}

	t.mu.Lock()
	t.remaining = rem
	t.resetAt = time.Unix(resetUnix, 0)
	t.mu.Unlock()
}

func isRetryableStatus(resp *http.Response) bool {
	if resp.StatusCode == http.StatusTooManyRequests {
		return true
	}
	if resp.StatusCode == http.StatusForbidden && isRateLimited(resp) {
		return true
	}
	return resp.StatusCode >= 500
}

func isRateLimited(resp *http.Response) bool {
	if resp.Header.Get("X-RateLimit-Remaining") == "0" {
		return true
	}
	return resp.Header.Get("Retry-After") != ""
}

func retryWait(resp *http.Response, attempt int) time.Duration {
	const maxWait = 120 * time.Second

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if seconds, err := strconv.Atoi(ra); err == nil && seconds > 0 {
				d := time.Duration(seconds) * time.Second
				if d > maxWait {
					return maxWait
				}
				return d
			}
		}
		if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
			if resetUnix, err := strconv.ParseInt(reset, 10, 64); err == nil {
				wait := time.Until(time.Unix(resetUnix, 0))
				if wait > 0 {
					if wait > maxWait {
						return maxWait
					}
					return wait
				}
			}
		}
		return 60 * time.Second
	}

	d := time.Duration(1<<uint(attempt)) * time.Second
	if d > maxWait {
		return maxWait
	}
	return d
}
