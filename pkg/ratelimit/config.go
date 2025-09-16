package ratelimit

import "math"

// DefaultBurstMultiplier defines how many seconds worth of tokens to allow,
// ensuring burst respects the specified rate limit without exceeding it.
const DefaultBurstMultiplier = 1

// Rate limiting configuration
type Config struct {
	// TxPerSecond limits the number of transactions per second
	TxPerSecond float64 `yaml:"tx_per_second,omitempty" json:"tx_per_second,omitempty"`

	// BytesPerSecond limits the transaction bytes processed per second
	BytesPerSecond uint64 `yaml:"bytes_per_second,omitempty" json:"bytes_per_second,omitempty"`

	// Burst configuration
	Burst BurstConfig `yaml:"burst" json:"burst"`
}

// BurstConfig defines burst allowance settings
type BurstConfig struct {
	// Size is the maximum number of tokens that can be consumed in a burst
	// For TxPerSecond, this is the max number of transactions
	// For BytesPerSecond, this is the max bytes
	// If Size > 0, burst is automatically enabled
	Size int `yaml:"size" json:"size"`
}

// IsEnabled returns true if any rate limiting is configured
func (c *Config) IsEnabled() bool {
	return c.TxPerSecond > 0 || c.BytesPerSecond > 0
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if !c.IsEnabled() {
		return nil
	}

	if c.TxPerSecond == 0 && c.BytesPerSecond == 0 {
		return ErrAtLeastOneRateLimitRequired
	}

	if c.TxPerSecond < 0 {
		return ErrTxPerSecondMustBePositive
	}

	return nil
}

// HasTxLimit returns true if transaction rate limiting is configured
func (c *Config) HasTxLimit() bool {
	return c.IsEnabled() && c.TxPerSecond > 0
}

// HasBytesLimit returns true if bytes rate limiting is configured
func (c *Config) HasBytesLimit() bool {
	return c.IsEnabled() && c.BytesPerSecond > 0
}

// GetTxRate returns the transaction rate
func (c *Config) GetTxRate() float64 {
	return c.TxPerSecond
}

// GetBytesRate returns the bytes rate
func (c *Config) GetBytesRate() float64 {
	return float64(c.BytesPerSecond)
}

// GetTpsBurstSize returns burst size specifically for TPS limiting
func (c *Config) GetTpsBurstSize() int {
	if c.Burst.Size > 0 {
		return c.Burst.Size
	}
	v := int(math.Ceil(c.TxPerSecond * float64(DefaultBurstMultiplier)))
	if v < 1 {
		return 1
	}
	return v
}

// GetBytesBurstSize returns burst size specifically for bytes limiting
func (c *Config) GetBytesBurstSize() int {
	if c.Burst.Size > 0 {
		return c.Burst.Size
	}
	// Saturating cast to int
	b := c.BytesPerSecond * uint64(DefaultBurstMultiplier)
	maxInt := int(^uint(0) >> 1)
	if b == 0 {
		return 1
	}
	if b > uint64(maxInt) {
		return maxInt
	}
	return int(b)
}
