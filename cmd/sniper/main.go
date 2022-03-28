package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"

	_ "net/http/pprof"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gelfand/evm-sniper/core"
	"github.com/gelfand/log"
)

var (
	privateKey   = flag.String("private.key", "", "wallet private key")
	poolAddr     = flag.String("pool.addr", "127.0.0.1:1111", "address of the txpool")
	serverAddr   = flag.String("server.addr", "127.0.0.1:2222", "address of the send-server")
	rpcAddr      = flag.String("rpc.addr", "http://127.0.0.1:8545", "address of the Ethereum RPC")
	contractAddr = flag.String("contract.addr", "", "address of the smart to use as router")
	tokenIn      = flag.String("token.in", "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", "token in for swap")
	tokenOut     = flag.String("token.out", "", "token out for swap")
	amountIn     = flag.String("amount.in", "100000000000000000", "amount of the input token")
	minLiquidity = flag.String("min.liquidity", "1000000000000000000", "minimum amount of the liquidity added")
	// Profiling.
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of sniper...\n")
	flag.PrintDefaults()
}

func main() { // nolint
	flag.Usage = usage
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Error("could not create CPU profile", "err", err)
			return
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Error("could not start CPU profile", "err", err)
			return
		}
		defer pprof.StopCPUProfile()
	}

	if *privateKey == "" {
		flag.Usage()
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	pk, err := crypto.HexToECDSA(*privateKey)
	if err != nil {
		log.Error("unable to parse private key", "err", err)
		return
	}

	amountInV, ok := new(big.Int).SetString(*amountIn, 10)
	if !ok {
		log.Info("Not okay! to set input amount")
		os.Exit(1)
	}
	minLiquidityV, ok := new(big.Int).SetString(*minLiquidity, 10)
	if !ok {
		log.Info("Not okay! to set input amount")
		os.Exit(1)
	}

	cfg := core.SniperConfig{
		TxPoolAddr:   *poolAddr,
		ServerAddr:   *serverAddr,
		RPCAddr:      *rpcAddr,
		ContractAddr: common.HexToAddress(*contractAddr),
		PrivateKey:   pk,
		TokenIn:      common.HexToAddress(*tokenIn),
		TokenOut:     common.HexToAddress(*tokenOut),
		AmountIn:     amountInV,
		MinLiquidity: minLiquidityV,
	}

	sniper, err := core.NewSniper(ctx, cfg)
	if err != nil {
		log.Error("Unable to create new Sniper", "err", err)
		return
	}
	sniper.Run(ctx)

	<-ctx.Done()

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Error("could not create memory profile", "err", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Error("could not write memory profile", "err", err)
		}
	}
}
