package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

// limiter wraps golang.org/x/time/rate.Limiter with our interface
type limiter struct {
	rateLimiter *rate.Limiter
}

func NewLimiter(ratePerSecond float64, burstSize int) (Limiter, error) {
	if ratePerSecond <= 0 {
		return nil, errors.Wrapf(ErrRateMustBePositive, "got %f", ratePerSecond)
	}
	if burstSize <= 0 {
		return nil, errors.Wrapf(ErrBurstSizeMustBePositive, "got %d", burstSize)
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
	GasLimit  uint64 // Gas limit for the transaction
}

// MultiLimiter wraps multiple types of rate limits
type MultiLimiter struct {
	tpsLimiter   Limiter
	bytesLimiter Limiter
	gasLimiter   Limiter
	txConfig     client.TxConfig // For decoding transactions to extract gas
	config       Config
}

// NewMultiLimiter creates a multi-metric rate limiter
func NewMultiLimiter(config Config, txConfig client.TxConfig) (*MultiLimiter, error) {
	ml := &MultiLimiter{
		config:   config,
		txConfig: txConfig,
	}

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

	if config.GasPerSecond > 0 {
		gasLimiter, err := NewLimiter(config.GetGasRate(), config.GetGasBurstSize())
		if err != nil {
			return nil, err
		}
		ml.gasLimiter = gasLimiter
	}

	return ml, nil
}

// WaitForTransactionBytes applies all configured rate limits for a transaction by decoding the bytes
func (ml *MultiLimiter) WaitForTransactionBytes(ctx context.Context, txBytes []byte) error {
	// Extract gas limit from transaction bytes
	gasLimit, err := ExtractGasLimit(ml.txConfig, txBytes)
	if err != nil {
		// If we can't extract gas, just use 0 and continue with other limits
		gasLimit = 0
	}

	metrics := TxMetrics{
		SizeBytes: uint64(len(txBytes)),
		GasLimit:  gasLimit,
	}

	return ml.WaitForTransaction(ctx, metrics)
}

// WaitForTransaction applies all configured rate limits for a transaction
func (ml *MultiLimiter) WaitForTransaction(ctx context.Context, txMetrics TxMetrics) error {
	// Apply TPS limiting (1 token per transaction)
	if ml.tpsLimiter != nil {
		if err := ml.tpsLimiter.Wait(ctx, 1); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}

			return errors.Wrap(ErrTPSRateLimitExceeded, err.Error())
		}
	}

	// Apply Bytes limiting (N tokens based on actual tx size)
	if ml.bytesLimiter != nil && txMetrics.SizeBytes > 0 {
		// Consume in chunks ≤ configured burst to avoid WaitN(n>burst) errors.
		remaining := txMetrics.SizeBytes
		chunk := ml.config.GetBytesBurstSize()
		for remaining > 0 {
			n := chunk
			if remaining < uint64(chunk) {
				n = int(remaining)
			}

			if err := ml.bytesLimiter.Wait(ctx, n); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return err
				}
				return errors.Wrapf(ErrBytesRateLimitExceeded, "failed to wait for bytes: %v", err)
			}

			remaining -= uint64(n)
		}
	}

	// Apply Gas limiting (N tokens based on actual gas limit)
	if ml.gasLimiter != nil && txMetrics.GasLimit > 0 {
		// Consume in chunks ≤ configured burst to avoid WaitN(n>burst) errors.
		remaining := txMetrics.GasLimit
		chunk := ml.config.GetGasBurstSize()
		for remaining > 0 {
			n := chunk
			if remaining < uint64(chunk) {
				n = int(remaining)
			}

			if err := ml.gasLimiter.Wait(ctx, n); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return err
				}

				return errors.Wrapf(ErrGasRateLimitExceeded, "failed to wait for gas: %v", err)
			}

			remaining -= uint64(n)
		}
	}

	return nil
}
