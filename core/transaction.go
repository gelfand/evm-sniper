package core

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// buildTransaction builds a transaction for the Ethereum EVM by given input params.
func buildTransaction(swpType byte, pair common.Address, amountIn *big.Int) []byte {
	buf := bytes.NewBuffer([]byte{swpType})
	buf.Write(pair.Bytes())
	buf.Write(amountIn.Bytes())

	return buf.Bytes()
}
