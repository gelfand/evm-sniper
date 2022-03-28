package etl

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var denominatorEther = big.NewInt(1e18)

// CalculatePairAddress calculates a pair address for the given factory and tokens.
func CalculatePairAddress(factory common.Address, token0 common.Address, token1 common.Address, salt []byte) common.Address {
	addrSum := append(token0.Bytes(), token1.Bytes()...)

	message := []byte{0xff}
	message = append(message, factory.Bytes()...)
	message = append(message, crypto.Keccak256(addrSum)...)
	message = append(message, salt...)

	hashed := crypto.Keccak256(message)

	addressBytes := new(big.Int).SetBytes(hashed)
	addressBytes = addressBytes.Abs(addressBytes)

	return common.BytesToAddress(addressBytes.Bytes())
}

func SortAddressess(tkn0 common.Address, tkn1 common.Address) (common.Address, common.Address, bool) {
	token0Rep := big.NewInt(0).SetBytes(tkn0.Bytes())
	token1Rep := big.NewInt(0).SetBytes(tkn1.Bytes())

	changed := false
	if token0Rep.Cmp(token1Rep) > 0 {
		tkn0, tkn1 = tkn1, tkn0
		changed = true
	}

	return tkn0, tkn1, changed
}

func ToEther(v int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(v), denominatorEther)
}
