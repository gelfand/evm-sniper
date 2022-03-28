package p2p

import (
	"context"
	"errors"
	"io"
	"net"
	"os"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/gelfand/log"
)

func DialContext(ctx context.Context, wcount int, serverAddr string, txChan <-chan *types.Transaction) {
	for w := 1; w < wcount; w++ {
		go worker(ctx, serverAddr, txChan)
	}
}

func worker(ctx context.Context, serverAddr string, txChan <-chan *types.Transaction) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Error("Unable to handle TCP connection", "err", err)
		os.Exit(1)
	}
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case tx := <-txChan:
			var txs types.Transactions
			txs = append(txs, tx)

			if err = rlp.Encode(conn, txs); err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				log.Error("Unable to encode transactions", "err", err)
			}
		}
	}
}
