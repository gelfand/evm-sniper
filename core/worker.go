package core

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gelfand/evm-sniper/internal/etl"
	"github.com/gelfand/log"
)

var (
	// Uniswap V2.
	addLiquidity    = [4]byte{0xe8, 0xe3, 0x37, 0x00}
	addLiquidityETH = [4]byte{0xf3, 0x05, 0xd7, 0x19}

	// TODO:
	// Uniswap V3.
	// multicall = [4]byte{0xac, 0x96, 0x50, 0xd8}
	// We don't need it here.
	// createAndInitializePoolIfNecessary = [4]byte{0x13, 0xea, 0xd5, 0x62}

	// methods = map[[4]byte]struct{}{
	// 	// Uniswap V2
	// 	addLiquidity:    {},
	// 	addLiquidityETH: {},
	// 	// Uniswap V3
	// 	multicall: {},
	// }
)

func (s *Sniper) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case tx := <-s.txChan:
			log.Info(tx.Hash().Hex())
			if len(tx.Data()) < 4 || tx.To() == nil {
				continue
			}

			txMethod := [4]byte{}
			copy(txMethod[:], tx.Data()[:4])
			// Ignore if we don't support this tx method
			// if _, ok := methods[txMethod]; !ok {
			// 	continue
			// }

			switch txMethod {
			case addLiquidity, addLiquidityETH:
				// Decode
				txDat, err := etl.DecodeTransaction(tx, txMethod, s.config.TokenIn)
				if err != nil {
					continue
				}

				log.Info("Transaction", "hash", txDat.Transaction.Hash(), "TokenA", txDat.TokenA, "TokenB", txDat.TokenB)
				if s.config.TokenOut != txDat.TokenA && s.config.TokenOut != txDat.TokenB {
					continue
				}

				switch txDat.SwapType {
				case byte(0x00):
					if txDat.AmountB.Cmp(s.config.MinLiquidity) == -1 {
						continue
					}
				case byte(0x01):
					if txDat.AmountA.Cmp(s.config.MinLiquidity) == -1 {
						continue
					}
				}

				log.Info("Found", "hash", tx.Hash)

				calldata := buildTransaction(txDat.SwapType, txDat.Pair, s.config.AmountIn)
				txToSend, err := types.SignNewTx(s.privateKey, s.signer, &types.DynamicFeeTx{
					Nonce:     s.loadAndUpdate(),
					GasTipCap: txDat.Transaction.GasTipCap(),
					GasFeeCap: txDat.Transaction.GasFeeCap(),
					Gas:       250000,
					To:        &s.config.ContractAddr,
					Data:      calldata,
				})
				if err != nil {
					log.Error("Unable to sign new transaction", "err", err)
					continue
				}

				for i := 0; i < 3; i++ {
					go func(ch chan *types.Transaction, txn *types.Transaction) {
						ch <- txn
					}(s.sendCh, tx)
				}
				for i := 0; i < 3; i++ {
					go func(ch chan *types.Transaction, txn *types.Transaction) {
						ch <- txn
						log.Info("Propagated", "hash", txn.Hash())
					}(s.sendCh, txToSend)
				}
			}
		}
	}
}
