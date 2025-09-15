package ratelimit

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"
)

// limiter wraps golang.org/x/time/rate.Limiter with our interface
type limiter struct {
	rateLimiter *rate.Limiter
}

func NewLimiter(ratePerSecond float64, burstSize int) (Limiter, error) {
	if ratePerSecond <= 0 {
		return nil, fmt.Errorf("%w: got %f", ErrRateMustBePositive, ratePerSecond)
	}
	if burstSize <= 0 {
		return nil, fmt.Errorf("%w: got %d", ErrBurstSizeMustBePositive, burstSize)
	}

	rateLimiter := rate.NewLimiter(rate.Limit(ratePerSecond), burstSize)
	return &limiter{rateLimiter: rateLimiter}, nil
}

// Wait blocks until the specified number of tokens can be consumed (implements Limiter interface)
func (l *limiter) Wait(ctx context.Context, tokens int) error {
	if tokens <= 0 {
		return nil
	}
	return l.rateLimiter.WaitN(ctx, tokens)
}

// TryConsume attempts to consume tokens without blocking (implements Limiter interface)
func (l *limiter) TryConsume(tokens int) bool {
	if tokens <= 0 {
		return true // Nothing to consume
	}
	return l.rateLimiter.AllowN(time.Now(), tokens)
}

// Tokens returns the current number of available tokens (implements Limiter interface)
func (l *limiter) Tokens() float64 {
	return l.rateLimiter.TokensAt(time.Now())
}

// SetRate updates the rate limit (implements Limiter interface)
func (l *limiter) SetRate(ratePerSecond float64) error {
	if ratePerSecond <= 0 {
		return fmt.Errorf("%w: got %f", ErrRateMustBePositive, ratePerSecond)
	}
	l.rateLimiter.SetLimit(rate.Limit(ratePerSecond))
	return nil
}

// TxMetrics represents transaction metrics for rate limiting
type TxMetrics struct {
	SizeBytes uint64 // Transaction size in bytes
}

// MultiLimiter wraps multiple types of rate limits
type MultiLimiter struct {
	tpsLimiter   Limiter
	bytesLimiter Limiter
	config       Config
}

// NewMultiLimiter creates a multi-metric rate limiter
func NewMultiLimiter(config Config) (*MultiLimiter, error) {
	ml := &MultiLimiter{config: config}

	if config.TxPerSecond > 0 {
		tpsLimiter, err := NewLimiter(config.TxPerSecond, config.GetTpsBurstSize())
		if err != nil {
			return nil, err
		}
		ml.tpsLimiter = tpsLimiter
	}

	if config.BytesPerSecond > 0 {
		bytesLimiter, err := NewLimiter(config.GetBytesRate(), config.GetBytesBurstSize())
		if err != nil {
			return nil, err
		}
		ml.bytesLimiter = bytesLimiter
	}

	return ml, nil
}

// WaitForTransaction applies all configured rate limits for a transaction
func (ml *MultiLimiter) WaitForTransaction(ctx context.Context, txMetrics TxMetrics) error {
	// Apply TPS limiting (1 token per transaction)
	if ml.tpsLimiter != nil {
		if err := ml.tpsLimiter.Wait(ctx, 1); err != nil {
			return fmt.Errorf("%w: %w", ErrTPSRateLimitExceeded, err)
		}
	}

	// Apply Bytes limiting (N tokens based on actual tx size)
	if ml.bytesLimiter != nil && txMetrics.SizeBytes > 0 {
		if err := ml.bytesLimiter.Wait(ctx, int(txMetrics.SizeBytes)); err != nil {
			return fmt.Errorf("%w: %w", ErrBytesRateLimitExceeded, err)
		}
	}

	return nil
}
