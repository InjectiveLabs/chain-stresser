package ratelimit

import (
	"context"
)

// BroadcastFunc represents a generic broadcast function signature
type BroadcastFunc func(ctx context.Context, txBytes []byte) (string, error)

// Limiter represents a rate limiter using the Token Bucket algorithm.
// This allows controlled bursts while maintaining an average rate over time.
type Limiter interface {
	// Wait blocks until the limiter permits the consumption of the specified number of tokens
	// Returns an error if the context is cancelled before permission is granted
	Wait(ctx context.Context, tokens int) error

	// TryConsume attempts to consume the specified number of tokens without blocking
	// Returns true if tokens were successfully consumed, false otherwise
	TryConsume(tokens int) bool

	// Tokens returns the current number of available tokens
	Tokens() float64

	// SetRate updates the rate limit (tokens per second)
	SetRate(rate float64) error
}
