package ratelimit

import (
	"math/rand"
	"testing"
	"time"
)

func TestCompact(t *testing.T) {
	now := time.Now()

	t.Run("RemoveOlderThan5Minutes", func(t *testing.T) {
		x := []event{
			{
				timeStamp: now.Add(-10 * time.Minute),
				latency:   0,
			},
			{
				timeStamp: now.Add(-4 * time.Minute),
				latency:   19 * time.Millisecond,
			},
		}

		r := rateLimitState{events: x}
		r.compact(now)

		if len(r.events) != 1 {
			t.Errorf("Expected %d, found %d\n", 1, len(r.events))
		}
	})

	t.Run("CompactEmpty", func(t *testing.T) {
		x := []event{}
		r := rateLimitState{events: x}
		r.compact(now)

		if len(r.events) != 0 {
			t.Errorf("Expected %d, found %d\n", 0, len(r.events))
		}
	})

	t.Run("CompactUnorderedByTime", func(t *testing.T) {
		x := []event{
			{
				timeStamp: now.Add(-4 * time.Minute),
				latency:   100 * time.Second,
			},
			{
				timeStamp: now.Add(-20 * time.Minute),
				latency:   4 * time.Millisecond,
			},
		}
		r := rateLimitState{events: x}
		r.compact(now)

		if len(r.events) != 1 {
			t.Fatalf("Expected %d, got %d\n", 1, len(r.events))
		}

		if r.events[0].latency != 100*time.Second {
			t.Errorf("Incorrect compaction")
		}
	})
}

func TestUpdateFraction(t *testing.T) {

	now := time.Now()
	t.Run("EmptyEvents", func(t *testing.T) {
		rateLimitState := NewRateLimitState(2*time.Second, "t")
		rateLimitState.updateFraction()
		if rateLimitState.fraction != 1 {
			t.Errorf("Expected %d, got %f\n", 1, rateLimitState.fraction)
		}
	})

	// Single event
	t.Run("SingleEvent_0.25", func(t *testing.T) {
		rateLimitState := NewRateLimitState(2*time.Second, "t")
		rateLimitState.update(10*time.Second, now)
		fraction := rateLimitState.getFraction()
		if fraction != 0.25 {
			t.Errorf("Expected %f, got %f\n", 0.25, fraction)
		}
	})

	t.Run("SingleEvent_1.0", func(t *testing.T) {
		rateLimitState := NewRateLimitState(2*time.Second, "t")
		rateLimitState.update(1*time.Second, now)
		fraction := rateLimitState.getFraction()
		if fraction != 1 {
			t.Errorf("Expected %f, got %f\n", 1., fraction)
		}
	})

	t.Run("Fraction0.5_sortedByLatency", func(t *testing.T) {
		rateLimitState := NewRateLimitState(2*time.Second, "t")

		for i := 0; i < 20; i++ {
			rateLimitState.update(time.Duration(i)*time.Second, now)
		}
		threshold := (rateLimitState.events[18].latency + rateLimitState.events[19].latency) / 2
		rateLimitState.latencyThreshold = threshold

		fraction := rateLimitState.getFraction()
		if fraction != 0.5 {
			t.Errorf("Expected %f, got %f\n", 0.5, fraction)
		}
	})

	t.Run("Fraction0.5_randomOrderOfEvents", func(t *testing.T) {
		rateLimitState := NewRateLimitState(2*time.Second, "t")

		for i := 0; i < 20; i++ {
			rateLimitState.update(time.Duration(i)*time.Second, now)
		}
		threshold := (rateLimitState.events[18].latency + rateLimitState.events[19].latency) / 2
		rateLimitState.latencyThreshold = threshold

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(rateLimitState.events), func(i, j int) {
			rateLimitState.events[i], rateLimitState.events[j] = rateLimitState.events[j], rateLimitState.events[i]
		})

		fraction := rateLimitState.getFraction()
		if fraction != 0.5 {
			t.Errorf("Expected %f, got %f\n", 0.5, fraction)
		}
	})
}

func TestUpdate(t *testing.T) {
	// Update calls compact
	now := time.Now()
	rateLimitState := NewRateLimitState(2*time.Second, "t")
	for i := 0; i < cap(rateLimitState.events); i++ {
		rateLimitState.update(1*time.Second, now)
	}

	if len(rateLimitState.events) != cap(rateLimitState.events) {
		t.Errorf("Expected len = %d, got = %d\n", cap(rateLimitState.events), len(rateLimitState.events))
	}

	rateLimitState.update(2*time.Second, now.Add(5*time.Minute))
	if len(rateLimitState.events) != 1 {
		t.Errorf("Expected len = %d, got = %d\n", 1, len(rateLimitState.events))
	}
}

func TestRecordLatency(t *testing.T) {
	limiter := New(2, 20, time.Second*3, "t")
	limiter.RecordLatency(10 * time.Second)
	if limiter.rateLimiter.Limit() != 2*0.25 {
		t.Errorf("Expected ratelimit = %f, got = %f\n", 0.5, limiter.rateLimiter.Limit())
	}
}
