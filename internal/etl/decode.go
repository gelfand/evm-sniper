package etl

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	addLiquidity    = [4]byte{0xe8, 0xe3, 0x37, 0x00}
	addLiquidityETH = [4]byte{0xf3, 0x05, 0xd7, 0x19}

	// TODO:
	// multicall       = [4]byte{0xac, 0x96, 0x50, 0xd8}

	// // Internal Uniswap V3 methods.
	// createAndInitializePoolIfNecessary = [4]byte{0x13, 0xea, 0xd5, 0x62}
)

var (
	errTooShort          = errors.New("calldata is too short")
	errUnsupportedMethod = errors.New("method is unsupported")
	ErrUnknownFactory    = errors.New("unknown factory")
	errIgnore            = errors.New("ignore")
)

type TransactionData struct {
	Transaction *types.Transaction

	AmountA *big.Int
	AmountB *big.Int

	Pair   common.Address
	TokenA common.Address
	TokenB common.Address

	SwapType byte
}

func DecodeTransaction(tx *types.Transaction, txMethod [4]byte, tokenIn common.Address) (TransactionData, error) {
	if len(tx.Data()) < 196 {
		return TransactionData{}, errTooShort
	}

	factoryDat, ok := Factories[*tx.To()]
	if !ok {
		return TransactionData{}, ErrUnknownFactory
	}

	switch txMethod {
	case addLiquidity:
		if len(tx.Data()) < 260 {
			return TransactionData{}, errTooShort
		}

		tokenA := common.BytesToAddress(tx.Data()[16:36])
		tokenB := common.BytesToAddress(tx.Data()[48:68])
		amountA := new(big.Int).SetBytes(tx.Data()[132:164])
		amountB := new(big.Int).SetBytes(tx.Data()[164:196])

		var changed bool
		tokenA, tokenB, changed = SortAddressess(tokenA, tokenB)
		if changed {
			amountA, amountB = amountB, amountA
		}

		var swpType byte
		switch tokenIn {
		case tokenA:
			swpType = 0x01
		case tokenB:
			swpType = 0x00
		default:
			return TransactionData{}, errIgnore
		}

		txData := TransactionData{
			Transaction: tx,
			TokenA:      tokenA,
			TokenB:      tokenB,
			AmountA:     amountA,
			AmountB:     amountB,
			SwapType:    swpType,
		}

		txData.Pair = CalculatePairAddress(factoryDat.Factory, txData.TokenA, txData.TokenB, factoryDat.Initcode)

		return txData, nil
	case addLiquidityETH:
		weth := WethAddr(tx.ChainId().Int64())

		tokenA := weth
		tokenB := common.BytesToAddress(tx.Data()[16:36])

		amountA := tx.Value()
		amountB := new(big.Int).SetBytes(tx.Data()[68:100])

		var changed bool
		tokenA, tokenB, changed = SortAddressess(tokenA, tokenB)
		if changed {
			amountA, amountB = amountB, amountA
		}

		var swpType byte
		switch weth {
		case tokenA:
			swpType = 0x01
		case tokenB:
			swpType = 0x00
		}

		txData := TransactionData{
			Transaction: tx,
			TokenA:      tokenA,
			TokenB:      tokenB,
			AmountA:     amountA,
			AmountB:     amountB,
			SwapType:    swpType,
		}
		txData.Pair = CalculatePairAddress(factoryDat.Factory, txData.TokenA, txData.TokenB, factoryDat.Initcode)
		return txData, nil
	default:
		return TransactionData{}, errUnsupportedMethod
	}
}
