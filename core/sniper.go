package core

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gelfand/evm-sniper/core/p2p"
	"github.com/gelfand/evm-sniper/core/txpool"
	"github.com/gelfand/log"
)

// SniperConfig are the configuration parameters for the Sniper.
type SniperConfig struct {
	// AmountIn is input amount for the swap.
	AmountIn *big.Int
	// MinLiquidity is a minimum liquidity added for triggering sniper.
	MinLiquidity *big.Int
	// PrivateKey is a wallet private key.
	PrivateKey *ecdsa.PrivateKey
	// TxPoolAddr is the address of the txpool server.
	TxPoolAddr string
	// ServerAddr is the address of the eth/handler.go
	// which is being used for fast direct propgation.
	ServerAddr string
	// RPCAddr is the address of the Ethereum RPC.
	RPCAddr string
	// ContractAddr is the address of the pseudo-router contract
	// which is mainly being used for sending transactions.
	ContractAddr common.Address
	// TokenIn is the input token.
	TokenIn common.Address
	// TokenOut is the output token.
	TokenOut common.Address
}

// Sniper is the sniper object for sniping liquidity
// and/or other events on the EVM-like blockchains.
type Sniper struct {
	config SniperConfig

	address    common.Address
	signer     types.Signer
	privateKey *ecdsa.PrivateKey
	ec         *ethclient.Client

	txChan chan *types.Transaction
	sendCh chan *types.Transaction
	// nonce is the wallet nonce.
	// BEING USED ONLY THROUGH ATOMICS!
	nonce uint64
}

func NewSniper(ctx context.Context, cfg SniperConfig) (*Sniper, error) {
	s := &Sniper{
		config:     cfg,
		address:    crypto.PubkeyToAddress(cfg.PrivateKey.PublicKey),
		privateKey: cfg.PrivateKey,
		txChan:     make(chan *types.Transaction),
		sendCh:     make(chan *types.Transaction),
		nonce:      0,
	}

	var err error
	s.ec, err = ethclient.DialContext(ctx, cfg.RPCAddr)
	if err != nil {
		return nil, fmt.Errorf("unable to establish connection with Ethereum RPC service: %w", err)
	}

	chainID, err := s.ec.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve chainID: %w", err)
	}

	s.signer = types.NewLondonSigner(chainID)

	nonce, err := s.ec.PendingNonceAt(ctx, s.address)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve wallet nonce: %w", err)
	}
	s.storeNonce(nonce)

	log.Info("Wallet", "address", s.address, "nonce", s.loadNonce())
	return s, nil
}

func (s *Sniper) Run(ctx context.Context) {
	txpool.DialContext(ctx, 8, s.config.TxPoolAddr, s.txChan)
	p2p.DialContext(ctx, 8, s.config.ServerAddr, s.sendCh)

	// Start sniper workers
	for w := 1; w <= 12; w++ {
		go s.worker(ctx)
	}
}

func (s *Sniper) loadAndUpdate() uint64 {
	return (atomic.AddUint64(&s.nonce, 1) - 1)
}

func (s *Sniper) storeNonce(val uint64) {
	atomic.StoreUint64(&s.nonce, val)
}

func (s *Sniper) loadNonce() uint64 {
	return atomic.LoadUint64(&s.nonce)
}
