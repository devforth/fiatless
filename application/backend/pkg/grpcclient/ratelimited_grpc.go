package grpcclient

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

type RateLimitedInterceptor struct {
	limiter *rate.Limiter
	mu      sync.RWMutex
}

func NewRateLimitedInterceptor(rps float64) *RateLimitedInterceptor {
	var limiter *rate.Limiter
	if rps > 0 {
		limiter = rate.NewLimiter(rate.Limit(rps), 1)
	}
	return &RateLimitedInterceptor{
		limiter: limiter,
	}
}

func (r *RateLimitedInterceptor) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		r.mu.RLock()
		limiter := r.limiter
		r.mu.RUnlock()

		if limiter != nil {
			if err := limiter.Wait(ctx); err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return fmt.Errorf("rate limit error: %v", err)
			}
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (r *RateLimitedInterceptor) UpdateRPS(rps float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rps <= 0 {
		r.limiter = nil
	} else {
		if r.limiter == nil {
			r.limiter = rate.NewLimiter(rate.Limit(rps), 1)
		} else {
			r.limiter.SetLimit(rate.Limit(rps))
		}
	}
}
