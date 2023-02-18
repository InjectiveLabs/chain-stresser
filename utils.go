package main

import (
	"math/big"
	"strings"

	cosmtypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	log "github.com/xlab/suplog"
)

func logLevel(s string) log.Level {
	switch s {
	case "1", "error":
		return log.ErrorLevel
	case "2", "warn":
		return log.WarnLevel
	case "3", "info":
		return log.InfoLevel
	case "4", "debug":
		return log.DebugLevel
	default:
		return log.FatalLevel
	}
}

// maxUInt256 returns a value equal to 2**256 - 1 (MAX_UINT in Solidity).
//
//nolint:all
func maxUInt256() *big.Int {
	return new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))
}

//nolint:all
func toBool(s string) bool {
	switch strings.ToLower(s) {
	case "true", "1", "t", "yes":
		return true
	default:
		return false
	}
}

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}

// price * 10^quoteDecimals/10^baseDecimals = price * 10^(quoteDecimals - baseDecimals)
// for INJ/USDT, INJ is the base which has 18 decimals and USDT is the quote which has 6 decimals
func getPrice(price float64, baseDecimals, quoteDecimals int) cosmtypes.Dec {
	scale := decimal.New(1, int32(quoteDecimals-baseDecimals))
	priceStr := scale.Mul(decimal.NewFromFloat(price)).String()
	decPrice, _ := cosmtypes.NewDecFromStr(priceStr)
	return decPrice
}

func defaultSubaccount(acc cosmtypes.AccAddress) common.Hash {
	return common.BytesToHash(common.RightPadBytes(acc.Bytes(), 32))
}

func amount(v float64, decimals ...int) cosmtypes.Dec {
	var decimalsOpt int32 = 18
	if len(decimals) == 1 {
		decimalsOpt = int32(decimals[0])
	} else if len(decimals) > 1 {
		panic("too much arguments")
	}

	d := decimal.NewFromFloat(v).Shift(decimalsOpt)
	return cosmtypes.NewDecFromInt(cosmtypes.NewIntFromBigInt(d.BigInt()))
}
