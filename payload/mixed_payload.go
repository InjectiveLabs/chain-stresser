package payload

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"os"
	"time"

	"github.com/pkg/errors"
	log "github.com/xlab/suplog"
	"gopkg.in/yaml.v3"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
)

type MixedPayloadConfig struct {
	StresserConfig      *StresserConfig            `yaml:"stresser_config,omitempty"`
	BankSend            *BankSendConfig            `yaml:"bank_send,omitempty"`
	BankMultiSend       *BankMultiSendConfig       `yaml:"bank_multi_send,omitempty"`
	EthSend             *EthSendConfig             `yaml:"eth_send,omitempty"`
	EthCall             *EthCallConfig             `yaml:"eth_call,omitempty"`
	EthDeploy           *EthDeployConfig           `yaml:"eth_deploy,omitempty"`
	ExchangeBatchOrders *ExchangeBatchOrdersConfig `yaml:"exchange_batch_orders,omitempty"`
	WasmStoreCode       *WasmStoreCodeConfig       `yaml:"wasm_store_code,omitempty"`
	WasmInitContract    *WasmInitContractConfig    `yaml:"wasm_init_contract,omitempty"`
	WasmExecContract    *WasmExecContractConfig    `yaml:"wasm_exec_contract,omitempty"`
}

type StresserConfig struct {
	ChainID             string  `yaml:"chain_id,omitempty"`
	EthChainID          int64   `yaml:"eth_chain_id,omitempty"`
	MinGasPrice         string  `yaml:"min_gas_price,omitempty"`
	NodeAddress         string  `yaml:"node_addr,omitempty"`
	GRPCAddress         string  `yaml:"grpc_addr,omitempty"`
	AwaitTxConfirmation *bool   `yaml:"await,omitempty"`
	Accounts            string  `yaml:"accounts,omitempty"`
	AccountsNum         int     `yaml:"accounts_num,omitempty"`
	NumOfTransactions   int     `yaml:"transactions,omitempty"`
	RateTPS             float64 `yaml:"rate_tps,omitempty"`
	RateBytes           uint64  `yaml:"rate_bytes,omitempty"`
	RateGas             uint64  `yaml:"rate_gas,omitempty"`
	RateBurstSize       int     `yaml:"rate_burst_size,omitempty"`
}

type BankSendConfig struct {
	Frequency  float64 `yaml:"frequency"`
	SendAmount string  `yaml:"send_amount,omitempty"`
}

type BankMultiSendConfig struct {
	Frequency  float64 `yaml:"frequency"`
	SendAmount string  `yaml:"send_amount,omitempty"`
	NumTargets int     `yaml:"num_targets,omitempty"`
}

type EthSendConfig struct {
	Frequency  float64 `yaml:"frequency"`
	SendAmount string  `yaml:"send_amount,omitempty"`
}

type EthCallConfig struct {
	Frequency float64 `yaml:"frequency"`
}

type EthDeployConfig struct {
	Frequency float64 `yaml:"frequency"`
}

type ExchangeBatchOrdersConfig struct {
	Frequency           float64  `yaml:"frequency"`
	SpotMarketIDs       []string `yaml:"spot_market_ids,omitempty"`
	DerivativeMarketIDs []string `yaml:"derivative_market_ids,omitempty"`
}

type WasmStoreCodeConfig struct {
	Frequency float64 `yaml:"frequency"`
}

type WasmInitContractConfig struct {
	Frequency float64 `yaml:"frequency"`
}

type WasmExecContractConfig struct {
	Frequency float64 `yaml:"frequency"`
}

func LoadMixedPayloadConfig(path string) (*MixedPayloadConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config file")
	}

	var cfg MixedPayloadConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Wrap(err, "failed to parse config file")
	}

	return &cfg, nil
}

type MixedPayloadProvider struct {
	providers   []TxProvider
	frequencies []float64
	rng         randSource
}

type randSource interface {
	Float64() float64
}

func NewMixedPayloadProvider(
	cfg *MixedPayloadConfig,
	minGasPrice string,
	ethChainID int64,
	queryClient chain.Client,
) (*MixedPayloadProvider, error) {
	var providers []TxProvider
	var frequencies []float64

	if cfg.BankSend != nil && cfg.BankSend.Frequency > 0 {
		sendAmount := cfg.BankSend.SendAmount
		if sendAmount == "" {
			sendAmount = "1" + chain.DefaultBondDenom
		}
		provider, err := NewBankSendProvider(minGasPrice, sendAmount)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create bank send provider")
		}
		providers = append(providers, provider)
		frequencies = append(frequencies, cfg.BankSend.Frequency)
	}

	if cfg.BankMultiSend != nil && cfg.BankMultiSend.Frequency > 0 {
		sendAmount := cfg.BankMultiSend.SendAmount
		if sendAmount == "" {
			sendAmount = "1" + chain.DefaultBondDenom
		}
		numTargets := cfg.BankMultiSend.NumTargets
		if numTargets == 0 {
			numTargets = 50
		}
		provider, err := NewBankMultiSendProvider(minGasPrice, sendAmount, numTargets)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create bank multi send provider")
		}
		providers = append(providers, provider)
		frequencies = append(frequencies, cfg.BankMultiSend.Frequency)
	}

	if cfg.EthSend != nil && cfg.EthSend.Frequency > 0 {
		sendAmount := cfg.EthSend.SendAmount
		if sendAmount == "" {
			sendAmount = "1" + chain.DefaultBondDenom
		}
		provider, err := NewEthSendProvider(big.NewInt(ethChainID), minGasPrice, sendAmount)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create eth send provider")
		}
		providers = append(providers, provider)
		frequencies = append(frequencies, cfg.EthSend.Frequency)
	}

	if cfg.EthCall != nil && cfg.EthCall.Frequency > 0 {
		provider, err := NewEthCallProvider(big.NewInt(ethChainID), minGasPrice)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create eth call provider")
		}
		providers = append(providers, provider)
		frequencies = append(frequencies, cfg.EthCall.Frequency)
	}

	if cfg.EthDeploy != nil && cfg.EthDeploy.Frequency > 0 {
		provider, err := NewEthDeployProvider(big.NewInt(ethChainID), minGasPrice)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create eth deploy provider")
		}
		providers = append(providers, provider)
		frequencies = append(frequencies, cfg.EthDeploy.Frequency)
	}

	if cfg.ExchangeBatchOrders != nil && cfg.ExchangeBatchOrders.Frequency > 0 {
		spotMarketIDs := cfg.ExchangeBatchOrders.SpotMarketIDs
		derivativeMarketIDs := cfg.ExchangeBatchOrders.DerivativeMarketIDs
		provider, err := NewExchangeBatchOrdersProvider(minGasPrice, spotMarketIDs, derivativeMarketIDs)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create exchange batch orders provider")
		}
		providers = append(providers, provider)
		frequencies = append(frequencies, cfg.ExchangeBatchOrders.Frequency)
	}

	if cfg.WasmStoreCode != nil && cfg.WasmStoreCode.Frequency > 0 {
		provider, err := NewWasmStoreCodeProvider(minGasPrice)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create wasm store code provider")
		}
		providers = append(providers, provider)
		frequencies = append(frequencies, cfg.WasmStoreCode.Frequency)
	}

	if cfg.WasmInitContract != nil && cfg.WasmInitContract.Frequency > 0 {
		provider, err := NewWasmInitContractProvider(queryClient, minGasPrice)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create wasm init contract provider")
		}
		providers = append(providers, provider)
		frequencies = append(frequencies, cfg.WasmInitContract.Frequency)
	}

	if cfg.WasmExecContract != nil && cfg.WasmExecContract.Frequency > 0 {
		provider, err := NewWasmExecContractProvider(queryClient, minGasPrice)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create wasm exec contract provider")
		}
		providers = append(providers, provider)
		frequencies = append(frequencies, cfg.WasmExecContract.Frequency)
	}

	if len(providers) == 0 {
		return nil, errors.New("no payload providers configured with frequency > 0")
	}

	normalizedFreqs := normalizeFrequencies(frequencies)

	providerNames := []string{}
	for _, provider := range providers {
		providerNames = append(providerNames, provider.Name())
	}

	providerFreqPairs := make([]string, len(providers))
	for i := range providers {
		providerFreqPairs[i] = fmt.Sprintf("%s=%.2f%%", providerNames[i], normalizedFreqs[i]*100)
	}

	log.WithFields(log.Fields{
		"providers": providerFreqPairs,
	}).Info("Mixed payload provider initialized")

	return &MixedPayloadProvider{
		providers:   providers,
		frequencies: normalizedFreqs,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func normalizeFrequencies(frequencies []float64) []float64 {
	total := 0.0
	for _, freq := range frequencies {
		total += freq
	}

	if total == 0 {
		return frequencies
	}

	normalized := make([]float64, len(frequencies))
	for i, freq := range frequencies {
		normalized[i] = math.Round(freq/total*10000) / 10000
	}

	return normalized
}

func (m *MixedPayloadProvider) selectProvider() TxProvider {
	r := m.rng.Float64()
	cumulative := 0.0

	for i, freq := range m.frequencies {
		cumulative += freq
		if r <= cumulative {
			return m.providers[i]
		}
	}

	return m.providers[len(m.providers)-1]
}

func (m *MixedPayloadProvider) Name() string {
	return "mixed-payload"
}

func (m *MixedPayloadProvider) GenerateInitialTx(req TxRequest) (Tx, error) {
	// IMPORTANT: We must call GenerateInitialTx for ALL providers, not just a random one.
	// Some providers (like EthCall) need to initialize state during GenerateInitialTx
	// (e.g., deploy contracts, store addresses) before GenerateTx can work.
	// Since we randomly select providers in GenerateTx, all providers must be initialized.

	// NOTE: This method is kept for backward compatibility but only returns the first initial tx.
	// Use GenerateAllInitialTxs() to get all initial txs from all providers.
	initialTxs, err := m.GenerateAllInitialTxs(req)
	if err != nil {
		return nil, err
	}

	if len(initialTxs) == 0 {
		return nil, nil
	}

	return initialTxs[0], nil
}

// GenerateAllInitialTxs generates initial transactions for ALL providers in the mix.
// This is necessary because some providers (like EthCall) store state during initialization
// (e.g., deployed contract addresses) that is later used in GenerateTx.
// Since GenerateTx randomly selects providers, all must be properly initialized.
func (m *MixedPayloadProvider) GenerateAllInitialTxs(req TxRequest) ([]Tx, error) {
	var initialTxs []Tx

	for _, provider := range m.providers {
		tx, err := provider.GenerateInitialTx(req)
		if err != nil {
			return nil, err
		}
		if tx != nil {
			initialTxs = append(initialTxs, tx.WithProvider(provider))
		}
	}

	return initialTxs, nil
}

func (m *MixedPayloadProvider) GenerateTx(req TxRequest) (Tx, error) {
	provider := m.selectProvider()
	tx, err := provider.GenerateTx(req)
	if err != nil {
		return nil, err
	}

	return tx.WithProvider(provider), nil
}

func (m *MixedPayloadProvider) BuildAndSignTx(
	client chain.Client,
	unsignedTx Tx,
) (Tx, error) {
	provider := unsignedTx.Provider()
	if provider == nil {
		return nil, errors.New("transaction has no associated provider")
	}

	return provider.BuildAndSignTx(client, unsignedTx)
}
