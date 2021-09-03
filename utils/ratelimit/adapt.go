package ratelimit

import (
	"sort"
	"sync"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

type rateLimitState struct {
	mu sync.Mutex

	latencyThreshold time.Duration // Threshold that controls fraction based on which quantile it falls in recorded events.
	fraction         float64       // rate limited to this fraction of the max rate.
	events           []event       // recorded events
	id               string
}

type event struct {
	timeStamp time.Time
	latency   time.Duration
}

func NewRateLimitState(latencyThreshold time.Duration, id string) *rateLimitState {
	return &rateLimitState{
		latencyThreshold: latencyThreshold,
		events:           make([]event, 0, 1000),
		fraction:         1,
		id:               id,
	}
}

func (state *rateLimitState) compact(now time.Time) {
	oldest := now.Add(-5 * time.Minute)
	newEvents := make([]event, 0, len(state.events))

	for _, elem := range state.events {
		if elem.timeStamp.After(oldest) {
			newEvents = append(newEvents, elem)
		}
	}

	if len(newEvents) == cap(newEvents) {

		ymlogger.LogCriticalf(state.id,
			"AdaptiveRateLimit Warning: compact couldn't remove events. This might increase memory usage. Current size %d\n",
			len(newEvents))
	}

	state.events = newEvents
}

// Updates the fraction according to the latency of the latest event.
func (state *rateLimitState) updateFraction() {
	if len(state.events) == 0 {
		return
	}

	sort.Slice(state.events, func(i, j int) bool { return state.events[i].latency < state.events[j].latency })

	p90Index := int(0.9 * float32(len(state.events)))
	p95Index := int(0.95 * float32(len(state.events)))

	p90Latency := state.events[p90Index].latency
	p95Latency := state.events[p95Index].latency

	// There are 3 possible relationships between threshold and 90th/95th percentile

	// 1. p90 <= p95 <= threshold => fraction = 1
	// 2. p90 <= threshold <= p95 => fraction = 0.5
	// 3. threshold <= p90 <= p95 => fraction = 0.25

	defer ymlogger.LogInfof(state.id,
		"ratelimit fraction: fraction=%f, p95=%s, p90=%s, numEvents=%d, threshold=%s\n",
		state.fraction, p95Latency, p90Latency, len(state.events), state.latencyThreshold)

	if p95Latency <= state.latencyThreshold {
		state.fraction = 1
		return
	}

	if p90Latency <= state.latencyThreshold {
		state.fraction = 0.5
		return
	}

	if p90Latency > state.latencyThreshold {
		state.fraction = 0.25
		return
	}
}

func (state *rateLimitState) update(latency time.Duration, now time.Time) {
	state.mu.Lock()
	defer state.mu.Unlock()

	if len(state.events)+1 > cap(state.events) {
		state.compact(now)
	}

	state.events = append(state.events, event{
		latency:   latency,
		timeStamp: now,
	})
}

func (state *rateLimitState) getFraction() float64 {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.updateFraction()
	return state.fraction
}
