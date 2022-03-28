package etl

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestDecodeTransaction(t *testing.T) {
	type args struct {
		tx       *types.Transaction
		txMethod [4]byte
		tokenIn  common.Address
	}
	tests := []struct {
		name    string
		args    args
		want    TransactionData
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeTransaction(tt.args.tx, tt.args.txMethod, tt.args.tokenIn)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
