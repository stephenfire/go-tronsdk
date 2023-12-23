package go_tronsdk

import (
	"crypto/sha256"

	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"google.golang.org/protobuf/proto"
)

func HashMessage(m proto.Message) ([]byte, error) {
	bs, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	hasher := sha256.New()
	hasher.Write(bs)
	return hasher.Sum(nil), nil
}

func ParseContractTxResult(tx *core.Transaction, runErr error) (fee int64, err error) {
	if runErr != nil {
		return 0, runErr
	}
	return (*Tx)(tx).ToResult()
}
