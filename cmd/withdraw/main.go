package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gelfand/evm-sniper/core/p2p"
)

var (
	privateKeyF   = flag.String("private.key", "", "wallet private key")
	serverAddr    = flag.String("server.addr", "127.0.0.1:2222", "address of the send-server")
	rpcAddr       = flag.String("rpc.addr", "http://127.0.0.1:8545", "address of the Ethereum RPC")
	contractAddrF = flag.String("contract.addr", "", "address of the smart to use as router")
	tokenAddr     = flag.String("token", "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", "token to withdraw")
)

func usage() {
	fmt.Fprintln(os.Stderr, "Usage of withdraw...")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	txChan := make(chan *types.Transaction)
	p2p.DialContext(ctx, 8, "127.0.0.1:2222", txChan)

	privateKey, err := crypto.HexToECDSA(*privateKeyF)
	if err != nil {
		log.Fatal(err)
	}
	client, err := ethclient.DialContext(ctx, *rpcAddr)
	if err != nil {
		log.Fatal(err)
	}

	from := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		log.Fatal(err)
	}

	tip, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		log.Fatal(err)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		log.Fatal(err)
	}

	signer := types.LatestSignerForChainID(chainID)
	contractAddr := common.HexToAddress(*contractAddrF)
	calldata := append([]byte{0x02}, common.FromHex(*tokenAddr)...)
	tx, err := types.SignNewTx(privateKey, signer, &types.DynamicFeeTx{
		Nonce:     nonce,
		Gas:       250000,
		GasTipCap: tip,
		GasFeeCap: big.NewInt(500e9),
		To:        &contractAddr,
		Data:      calldata,
	})
	if err != nil {
		log.Fatal(err)
	}
	txChan <- tx

	done := make(chan bool)
	go watch(ctx, client, tx.Hash(), done)

	isDone := <-done
	if !isDone {
		log.Println("Haven't waited for a full transaction confirmation")
		return
	}

	log.Printf("Successfully withdrawn, hash: %x", tx.Hash())
}

func watch(ctx context.Context, client *ethclient.Client, hash common.Hash, done chan bool) {
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			done <- false
			return
		case <-ticker.C:
			if _, _, err := client.TransactionByHash(ctx, hash); err == nil {
				done <- true
				return
			}
		}
	}
}
