package ratelimit

import "errors"

// Sentinel errors for rate limiting operations
var (
	// ErrRateMustBePositive indicates an invalid rate parameter
	ErrRateMustBePositive = errors.New("rate must be positive")

	// ErrBurstSizeMustBePositive indicates an invalid burst size parameter
	ErrBurstSizeMustBePositive = errors.New("burst size must be positive")

	// ErrAtLeastOneRateLimitRequired indicates no rate limits were configured
	ErrAtLeastOneRateLimitRequired = errors.New("at least one rate limit must be configured when enabled")

	// ErrTxPerSecondMustBePositive indicates an invalid TPS rate
	ErrTxPerSecondMustBePositive = errors.New("tx_per_second must be positive")

	// ErrBytesPerSecondMustBePositive indicates an invalid bytes rate
	ErrBytesPerSecondMustBePositive = errors.New("bytes_per_second must be positive")

	// ErrTPSRateLimitExceeded indicates TPS rate limit was hit
	ErrTPSRateLimitExceeded = errors.New("TPS rate limit exceeded")

	// ErrBytesRateLimitExceeded indicates bytes rate limit was hit
	ErrBytesRateLimitExceeded = errors.New("bytes rate limit exceeded")
)
