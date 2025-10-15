package httpclient

import (
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimitedTransport wraps an http.RoundTripper with rate limiting.
type RateLimitedTransport struct {
	baseTransport http.RoundTripper
	limiter       *rate.Limiter // Pointer, can be nil
	mu            sync.RWMutex
}

// NewRateLimitedTransport creates a new transport with rate limiting.
// If rps <= 0, rate limiting is disabled.
// If base is nil, http.DefaultTransport is used.
func NewRateLimitedTransport(rps float64, base http.RoundTripper) *RateLimitedTransport {
	if base == nil {
		base = http.DefaultTransport
	}

	rt := &RateLimitedTransport{
		baseTransport: base,
	}

	if rps > 0 {
		rt.limiter = rate.NewLimiter(rate.Limit(rps), 1) // Burst size 1
	}

	return rt
}

// RoundTrip executes a single HTTP transaction, applying rate limiting if enabled.
func (t *RateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	t.mu.RLock()
	limiter := t.limiter
	t.mu.RUnlock()

	// Apply rate limit only if limiter exists (is not nil)
	if limiter != nil {
		if err := limiter.Wait(ctx); err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, fmt.Errorf("rate limit error waiting for token: %w", err)
		}
	}

	return t.baseTransport.RoundTrip(req)
}

// UpdateRPS updates the rate limit.
// If rps <= 0, rate limiting is disabled.
func (t *RateLimitedTransport) UpdateRPS(rps float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if rps <= 0 {
		t.limiter = nil
	} else {
		if t.limiter == nil {
			t.limiter = rate.NewLimiter(rate.Limit(rps), 1)
		} else {
			t.limiter.SetLimit(rate.Limit(rps))
		}
	}
}

// CancelRequest cancels an in-flight request by closing its connection.
// Implements the optional CancelRequest method from http.RoundTripper.
func (t *RateLimitedTransport) CancelRequest(req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}
	if cr, ok := t.baseTransport.(canceler); ok {
		cr.CancelRequest(req)
	}
}
