package txpool

import (
	"context"
	"errors"
	"io"
	"net"
	"os"

	"github.com/gelfand/evm-sniper/internal/cbor"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gelfand/log"
)

func worker(ctx context.Context, poolAddr string, txChan chan *types.Transaction) {
	conn, err := net.Dial("tcp", poolAddr)
	if err != nil {
		log.Error("Unable to start txpool dialer", "err", err)
		os.Exit(1)
	}
	defer conn.Close()

	dec := cbor.Decoder(conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var tx *types.Transaction
		if err = dec.Decode(&tx); err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			log.Error("Unable to decode transaction", "err", err)
			continue
		}
		txChan <- tx
	}
}

func DialContext(ctx context.Context, wcount int, poolAddr string, txChan chan *types.Transaction) {
	for w := 1; w <= wcount; w++ {
		go worker(ctx, poolAddr, txChan)
	}
}
