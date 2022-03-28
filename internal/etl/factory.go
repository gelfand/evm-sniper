package etl

import "github.com/ethereum/go-ethereum/common"

type AMMData struct {
	Router   common.Address
	Factory  common.Address
	Initcode []byte
}

var Factories = map[common.Address]AMMData{
	// Uniswap V2
	common.HexToAddress("0x7a250d5630b4cf539739df2c5dacb4c659f2488d"): {
		Router:   common.HexToAddress("0x7a250d5630b4cf539739df2c5dacb4c659f2488d"),
		Factory:  common.HexToAddress("0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f"),
		Initcode: common.FromHex("96e8ac4277198ff8b6f785478aa9a39f403cb768dd02cbee326c3e7da348845f"),
	},
	// Sushiswap on Goerli
	common.HexToAddress("0x1b02da8cb0d097eb8d57a175b88c7d8b47997506"): {
		Router:   common.HexToAddress("0x1b02da8cb0d097eb8d57a175b88c7d8b47997506"),
		Factory:  common.HexToAddress("0xc35dadb65012ec5796536bd9864ed8773abc74c4"),
		Initcode: common.FromHex("e18a34eb0e04b04f7a0ac29a6e80748dca96319b42c54d679cb821dca90c6303"),
	},
	// Sushiswap on Mainnet
	common.HexToAddress("0xd9e1cE17f2641f24aE83637ab66a2cca9C378B9F"): {
		Router:   common.HexToAddress("0xd9e1cE17f2641f24aE83637ab66a2cca9C378B9F"),
		Factory:  common.HexToAddress("0xC0AEe478e3658e2610c5F7A4A2E1777cE9e4f2Ac"),
		Initcode: common.FromHex("e18a34eb0e04b04f7a0ac29a6e80748dca96319b42c54d679cb821dca90c6303"),
	},
}
