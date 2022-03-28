package etl

import "github.com/ethereum/go-ethereum/common"

func WethAddr(chainID int64) common.Address {
	switch chainID {
	case 1:
		return common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2")
	case 5:
		return common.HexToAddress("0xb4fbf271143f4fbf7b91a5ded31805e42b2208d6")
	default:
		return common.Address{}
	}
}
