// Package ratelimit provides transaction rate limiting using golang.org/x/time/rate.
//
// Supports TPS and bandwidth limiting with Token Bucket algorithm for controlled bursts.
//
// Basic Usage:
//
//	config := ratelimit.Config{
//		TxPerSecond:    100,   // 100 TPS limit
//		BytesPerSecond: 50000, // 50KB/sec limit
//	}
//
//	limiter, err := ratelimit.NewMultiLimiter(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// For each transaction:
//	txMetrics := ratelimit.TxMetrics{
//		SizeBytes: uint64(len(txBytes)),
//	}
//	err = limiter.WaitForTransaction(ctx, txMetrics)
//
// Custom burst sizes can be set via Config.Burst.Size.
// Default burst allows 1x the rate (e.g., 100 TPS = 100 burst).
package ratelimit
