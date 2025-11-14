package payload

import (
	"math"
	"math/rand"
	"testing"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
)

func TestNormalizeFrequencies(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected []float64
	}{
		{
			name:     "equal frequencies",
			input:    []float64{0.5, 0.5},
			expected: []float64{0.5, 0.5},
		},
		{
			name:     "unequal frequencies",
			input:    []float64{0.3, 0.7},
			expected: []float64{0.3, 0.7},
		},
		{
			name:     "frequencies that don't sum to 1",
			input:    []float64{1.0, 2.0, 3.0},
			expected: []float64{1.0 / 6.0, 2.0 / 6.0, 3.0 / 6.0},
		},
		{
			name:     "single frequency",
			input:    []float64{1.0},
			expected: []float64{1.0},
		},
		{
			name:     "multiple equal frequencies",
			input:    []float64{1.0, 1.0, 1.0, 1.0},
			expected: []float64{0.25, 0.25, 0.25, 0.25},
		},
		{
			name:     "very small frequencies",
			input:    []float64{0.01, 0.02, 0.03},
			expected: []float64{0.01 / 0.06, 0.02 / 0.06, 0.03 / 0.06},
		},
		{
			name:     "all zeros",
			input:    []float64{0.0, 0.0, 0.0},
			expected: []float64{0.0, 0.0, 0.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeFrequencies(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
				return
			}

			for i := range result {
				if math.Abs(result[i]-tt.expected[i]) > 1e-9 {
					t.Errorf("index %d: expected %f, got %f", i, tt.expected[i], result[i])
				}
			}

			sum := 0.0
			for _, freq := range result {
				sum += freq
			}

			if tt.name != "all zeros" && math.Abs(sum-1.0) > 1e-9 {
				t.Errorf("frequencies should sum to 1.0, got %f", sum)
			}
		})
	}
}

func TestSelectProvider(t *testing.T) {
	mockProvider1 := &mockTxProvider{name: "provider1"}
	mockProvider2 := &mockTxProvider{name: "provider2"}
	mockProvider3 := &mockTxProvider{name: "provider3"}

	tests := []struct {
		name        string
		providers   []interface{}
		frequencies []float64
		iterations  int
		tolerance   float64
	}{
		{
			name:        "equal distribution 50/50",
			providers:   []interface{}{mockProvider1, mockProvider2},
			frequencies: []float64{0.5, 0.5},
			iterations:  10000,
			tolerance:   0.05,
		},
		{
			name:        "unequal distribution 70/30",
			providers:   []interface{}{mockProvider1, mockProvider2},
			frequencies: []float64{0.7, 0.3},
			iterations:  10000,
			tolerance:   0.05,
		},
		{
			name:        "three providers equal",
			providers:   []interface{}{mockProvider1, mockProvider2, mockProvider3},
			frequencies: []float64{1.0 / 3.0, 1.0 / 3.0, 1.0 / 3.0},
			iterations:  15000,
			tolerance:   0.05,
		},
		{
			name:        "three providers weighted",
			providers:   []interface{}{mockProvider1, mockProvider2, mockProvider3},
			frequencies: []float64{0.5, 0.3, 0.2},
			iterations:  10000,
			tolerance:   0.05,
		},
		{
			name:        "extreme distribution 95/5",
			providers:   []interface{}{mockProvider1, mockProvider2},
			frequencies: []float64{0.95, 0.05},
			iterations:  10000,
			tolerance:   0.02,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counts := make(map[string]int)
			for i := range tt.providers {
				counts[tt.providers[i].(*mockTxProvider).name] = 0
			}

			mpp := &MixedPayloadProvider{
				providers:   convertMockProviders(tt.providers),
				frequencies: tt.frequencies,
				rng:         rand.New(rand.NewSource(12345)),
			}

			for i := 0; i < tt.iterations; i++ {
				provider := mpp.selectProvider()
				name := provider.Name()
				counts[name]++
			}

			for i, expectedFreq := range tt.frequencies {
				providerName := tt.providers[i].(*mockTxProvider).name
				actualCount := counts[providerName]
				actualFreq := float64(actualCount) / float64(tt.iterations)
				diff := math.Abs(actualFreq - expectedFreq)

				if diff > tt.tolerance {
					t.Errorf("provider %s: expected frequency %.3f, got %.3f (diff: %.3f, tolerance: %.3f)",
						providerName, expectedFreq, actualFreq, diff, tt.tolerance)
				}
			}

			totalCount := 0
			for _, count := range counts {
				totalCount += count
			}
			if totalCount != tt.iterations {
				t.Errorf("total count %d does not match iterations %d", totalCount, tt.iterations)
			}
		})
	}
}

func TestSelectProviderBoundaries(t *testing.T) {
	mockProvider1 := &mockTxProvider{name: "provider1"}
	mockProvider2 := &mockTxProvider{name: "provider2"}

	mpp := &MixedPayloadProvider{
		providers:   convertMockProviders([]interface{}{mockProvider1, mockProvider2}),
		frequencies: []float64{0.5, 0.5},
		rng:         &deterministicRNG{values: []float64{0.0, 0.49, 0.50, 0.51, 0.99, 1.0}},
	}

	expected := []string{"provider1", "provider1", "provider1", "provider2", "provider2", "provider2"}

	for i, exp := range expected {
		result := mpp.selectProvider()
		if result.Name() != exp {
			t.Errorf("iteration %d (r=%.2f): expected %s, got %s", i, mpp.rng.(*deterministicRNG).values[i], exp, result.Name())
		}
	}
}

type mockTxProvider struct {
	name string
}

func (m *mockTxProvider) Name() string {
	return m.name
}

func (m *mockTxProvider) GenerateInitialTx(req TxRequest) (Tx, error) {
	return nil, nil
}

func (m *mockTxProvider) GenerateTx(req TxRequest) (Tx, error) {
	return nil, nil
}

func (m *mockTxProvider) BuildAndSignTx(client chain.Client, unsignedTx Tx) (Tx, error) {
	return nil, nil
}

func convertMockProviders(mocks []interface{}) []TxProvider {
	result := make([]TxProvider, len(mocks))
	for i, mock := range mocks {
		result[i] = mock.(*mockTxProvider)
	}
	return result
}

type deterministicRNG struct {
	values []float64
	index  int
}

func (d *deterministicRNG) Float64() float64 {
	val := d.values[d.index%len(d.values)]
	d.index++
	return val
}
