package ratelimit

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

// A simple adaptive ratelimiter which uses rate limiter provided by go library and changes the rate
// based on latency metric. The logic is to use a threshold and it's relationship with 90th and 95th
// percentiles of recorded latencies and decide one of the 3 rates (can be more if we extend the logic of
// comparing the threshold to other percentiles).
// The implementation uses a fraction corresponding to how threshold is related to the latency percentiles
// and multiplies it with max rate to get instantaneous rate.

type AdaptiveRateLimiter struct {
	maxRate        float64
	rateLimiter    *rate.Limiter
	rateLimitState *rateLimitState
}

// Creates a new instance of adaptive rate limiter. This should be created for each independent scope
// of adaptation. For example, if we need to adapt to a bot's performance, this should be created per
// bot instance.
// maxRate is the maximum rate of api calls allowed. burst indicates maximum number of api calls that are allowed simultaneously
// by this rate limiter.
func New(maxRate float64, burst int, threshold time.Duration, id string) *AdaptiveRateLimiter {
	result := &AdaptiveRateLimiter{
		maxRate,
		rate.NewLimiter(rate.Limit(maxRate), burst),
		nil,
	}

	result.rateLimitState = NewRateLimitState(threshold, id)
	return result
}

// Based on the recorded latencies, the rate limit is adapted.
func (limit *AdaptiveRateLimiter) RecordLatency(latency time.Duration) {
	limit.rateLimitState.update(latency, time.Now())
	limit.rateLimiter.SetLimit(rate.Limit(limit.rateLimitState.getFraction() * limit.maxRate))
}

// Blocks until it's ok to invoke the service API based on configured ratelimit.
func (limit *AdaptiveRateLimiter) Wait(ctx context.Context) {
	limit.rateLimiter.Wait(ctx)
}
