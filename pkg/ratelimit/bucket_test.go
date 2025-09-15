package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	t.Run("NewLimiter validates parameters", func(t *testing.T) {
		tests := []struct {
			name      string
			rate      float64
			burstSize int
			wantErr   bool
		}{
			{"valid parameters", 10.0, 20, false},
			{"zero rate", 0, 20, true},
			{"negative rate", -1.0, 20, true},
			{"zero burst size", 10.0, 0, true},
			{"negative burst size", 10.0, -1, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := NewLimiter(tt.rate, tt.burstSize)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewLimiter() error = %v, wantErr %v", err, tt.wantErr)
				}

				// Test specific error types
				if tt.rate <= 0 && err != nil {
					if !errors.Is(err, ErrRateMustBePositive) {
						t.Errorf("Expected ErrRateMustBePositive, got %v", err)
					}
				}
				if tt.burstSize <= 0 && tt.rate > 0 && err != nil {
					if !errors.Is(err, ErrBurstSizeMustBePositive) {
						t.Errorf("Expected ErrBurstSizeMustBePositive, got %v", err)
					}
				}
			})
		}
	})

	t.Run("Wait with context cancellation", func(t *testing.T) {
		limiter, err := NewLimiter(1.0, 1) // 1 TPS, burst 1
		if err != nil {
			t.Fatal(err)
		}

		// First wait should succeed immediately
		ctx := context.Background()
		if err := limiter.Wait(ctx, 1); err != nil {
			t.Errorf("First wait failed: %v", err)
		}

		// Second wait should block, test cancellation
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		if err := limiter.Wait(ctx, 1); err == nil {
			t.Error("Expected error from cancelled context")
		}
	})

	t.Run("SetRate updates rate", func(t *testing.T) {
		limiter, err := NewLimiter(10.0, 10)
		if err != nil {
			t.Fatal(err)
		}

		// Test valid rate update
		if err := limiter.SetRate(20.0); err != nil {
			t.Errorf("SetRate failed: %v", err)
		}

		// Test invalid rate
		if err := limiter.SetRate(-1.0); err == nil {
			t.Error("Expected error for negative rate")
		}
	})
}

func TestMultiLimiter(t *testing.T) {
	t.Run("NewMultiLimiter with TPS only", func(t *testing.T) {
		config := Config{
			TxPerSecond: 10.0,
			Burst: BurstConfig{
				Size: 20,
			},
		}

		limiter, err := NewMultiLimiter(config)
		if err != nil {
			t.Fatal(err)
		}

		if limiter == nil {
			t.Error("Expected non-nil limiter")
		}

		// Test TPS limiting
		ctx := context.Background()
		txMetrics := TxMetrics{} // Empty metrics for TPS-only

		if err := limiter.WaitForTransaction(ctx, txMetrics); err != nil {
			t.Errorf("WaitForTransaction failed: %v", err)
		}
	})

	t.Run("NewMultiLimiter with TPS and bytes limits", func(t *testing.T) {
		config := Config{
			TxPerSecond:    10.0,
			BytesPerSecond: 50000,
			// Let burst size auto-calculate (default multiplier applied)
		}

		limiter, err := NewMultiLimiter(config)
		if err != nil {
			t.Fatal(err)
		}

		if limiter == nil {
			t.Error("Expected non-nil limiter")
		}

		// Test with real transaction metrics
		ctx := context.Background()
		txMetrics := TxMetrics{
			SizeBytes: 150,
		}

		if err := limiter.WaitForTransaction(ctx, txMetrics); err != nil {
			t.Errorf("WaitForTransaction with metrics failed: %v", err)
		}
	})

	t.Run("NewMultiLimiter with no limits", func(t *testing.T) {
		config := Config{} // No limits configured

		limiter, err := NewMultiLimiter(config)
		if err != nil {
			t.Fatal(err)
		}

		// Should work without any limiting
		ctx := context.Background()
		txMetrics := TxMetrics{}

		if err := limiter.WaitForTransaction(ctx, txMetrics); err != nil {
			t.Errorf("WaitForTransaction with no limits failed: %v", err)
		}
	})
}

func TestPerformanceBasic(t *testing.T) {
	// Just a basic smoke test to ensure our wrapper doesn't add significant overhead
	t.Run("Limiter performance", func(t *testing.T) {
		limiter, err := NewLimiter(1000.0, 1000) // High rate for testing
		if err != nil {
			t.Fatal(err)
		}

		start := time.Now()
		ctx := context.Background()

		// Try to consume 100 tokens quickly
		for i := 0; i < 100; i++ {
			if err := limiter.Wait(ctx, 1); err != nil {
				t.Fatalf("Wait failed at iteration %d: %v", i, err)
			}
		}

		elapsed := time.Since(start)
		if elapsed > 500*time.Millisecond {
			t.Errorf("Performance test took too long: %v", elapsed)
		}
	})
}
