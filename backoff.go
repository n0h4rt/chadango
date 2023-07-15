package chadango

import (
	"context"
	"math/rand"
	"time"
)

// Backoff represents a backoff mechanism with configurable duration and maximum duration.
type Backoff struct {
	Duration    time.Duration      // Duration represents the current backoff duration.
	MaxDuration time.Duration      // MaxDuration is the maximum allowed backoff duration.
	context     context.Context    // context is the context used for cancellation.
	cancel      context.CancelFunc // cancel is the function to cancel the `Backoff.Sleep()` operation.
}

// increment increases the backoff duration using an exponential strategy
func (b *Backoff) increment() {
	if b.Duration < b.MaxDuration {
		// Use an exponential backoff strategy to increase the wait duration between reconnection attempts.
		b.Duration *= 2
	}

	if b.Duration > b.MaxDuration {
		b.Duration = b.MaxDuration
	}
}

// Sleep is a mock of time.Sleep(), that is also responsive to the cancel signal.
// It adds some jitter to the context and waits until the context is done.
// Returns true if the sleep was cancelled before the deadline, false otherwise.
func (b *Backoff) Sleep(ctx context.Context) bool {
	defer b.increment()

	// Add some jitter to the context.
	b.context, b.cancel = context.WithTimeout(ctx, b.Duration+time.Duration(rand.Int63n(int64(b.Duration)/4)))
	defer b.cancel()

	<-b.context.Done()

	return b.context.Err() != context.DeadlineExceeded
}

// Cancel cancels the ongoing backoff sleep.
func (b *Backoff) Cancel() {
	b.cancel()
}
