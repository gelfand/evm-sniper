package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gelfand/log"
)

var (
	path        = flag.String("path", "./contracts/main.yul", "path to the contract")
	deploy      = flag.Bool("deploy", false, "deploy after compiling")
	rpcAddr     = flag.String("rpc.addr", "http://127.0.0.1:8545", "Ethereum RPC address")
	privateKeyV = flag.String("private.key", "", "wallet private key, required only if deploy=true")
)

func usage() {
	fmt.Fprintln(os.Stderr, "Usage of compile...")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	contractPath, err := filepath.Abs(*path)
	if err != nil {
		log.Error("Unable to resolve file path", "err", err)
		flag.Usage()
	}

	binString, err := compile(contractPath)
	if err != nil {
		log.Error("Unable to compile contract", "err", err)
	}

	if !(*deploy) {
		// Print and exit.
		fmt.Println(binString)
		return
	}

	if *privateKeyV == "" {
		log.Error("Unable to deploy contract", "err", "empty private key")
		fmt.Printf("Contract bin:\n %s\n\n", binString)
		flag.Usage()
	}

	bin := common.FromHex(binString)
	privateKey, err := crypto.HexToECDSA(*privateKeyV)
	if err != nil {
		log.Error("Invalid private key", "err", err)
		flag.Usage()
	}

	client, err := ethclient.DialContext(ctx, *rpcAddr)
	if err != nil {
		log.Error("Unable to establish connection with Ethereum RPC", "err", err)
		flag.Usage()
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		log.Error("Unable to retrieve chainID", "err", err)
		return
	}

	account := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, account)
	if err != nil {
		log.Error("Unable to retrieve wallet nonce", "err", err)
		return
	}

	signer := types.NewLondonSigner(chainID)
	tx, err := types.SignNewTx(privateKey, signer, &types.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(150e9), // 150 gwei
		Gas:      1_000_000,
		Data:     bin,
	})
	if err != nil {
		log.Error("Unable to create and sign new transaction", "err", err)
		flag.Usage()
	}

	if err = client.SendTransaction(ctx, tx); err != nil {
		log.Error("Unable to send transaction", "err", err)
		return
	}

	contractAddr := crypto.CreateAddress(account, nonce)
	log.Info("Successfully submitted new contract deployment", "hash", tx.Hash(), "contractAddr", contractAddr)
}

func compile(contractPath string) (string, error) {
	out, err := exec.Command("solc", "--strict-assembly", "--optimize", contractPath).Output()
	if err != nil {
		return "", fmt.Errorf("Unable to compile contract: %w", err)
	}

	bin := strings.SplitAfter(string(out), "Binary representation:\n")
	if len(bin) < 1 {
		return "", fmt.Errorf("Unable to parse output")
	}

	return (strings.Split(bin[1], "\n")[0]), nil
}
