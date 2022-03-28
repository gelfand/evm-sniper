package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	privateKey, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	wallet := crypto.PubkeyToAddress(privateKey.PublicKey)
	db := rawdb.NewMemoryDatabase()

	// Ensure that key1 has some funds in the genesis block.
	gspec := &core.Genesis{
		Config: &params.ChainConfig{HomesteadBlock: new(big.Int)},
		Alloc:  core.GenesisAlloc{wallet: {Balance: big.NewInt(5e18)}},
	}
	genesis := gspec.MustCommit(db)

	// This call generates a chain of 5 blocks. The function runs for
	// each block and adds different features to gen based on the
	// block index.
	signer := types.HomesteadSigner{}
	chain, _ := core.GenerateChain(gspec.Config, genesis, ethash.NewFaker(), db, 5, func(i int, gen *core.BlockGen) {
		switch i {
		case 0:
			// In block 1, addr1 sends addr2 some ether.
			tx, _ := types.SignTx(types.NewTransaction(gen.TxNonce(wallet), wallet, big.NewInt(10000), params.TxGas, nil, nil), signer, privateKey)
			gen.AddTx(tx)
		}
	})

	// Import the chain. This runs all block validation rules.
	blockchain, _ := core.NewBlockChain(db, nil, gspec.Config, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(chain); err != nil {
		fmt.Printf("insert error (block %d): %v\n", chain[i].NumberU64(), err)
		return
	}

	state, _ := blockchain.State()
	fmt.Printf("last block: #%d\n", blockchain.CurrentBlock().Number())
	fmt.Println("balance of addr3:", state.GetBalance(wallet))

	fmt.Printf("%+v\n", blockchain.Config())

	<-ctx.Done()
}
