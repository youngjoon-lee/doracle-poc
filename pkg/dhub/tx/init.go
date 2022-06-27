package tx

import (
	"fmt"

	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	log "github.com/sirupsen/logrus"
	oracletypes "github.com/youngjoon-lee/dhub/x/oracle/types"
)

func (e Executor) Init(operatorAddress string, enclaveReport []byte, oraclePubKey *secp256k1.PubKey) error {
	msg := oracletypes.NewMsgInit(operatorAddress, enclaveReport, oraclePubKey)

	res, err := e.signAndBroadcastTx(msg)
	if err != nil {
		return fmt.Errorf("failed to sign and broadcast tx: %w", err)
	}
	log.Debugf("tx res:%v", res)
	if res.Code != 0 {
		return fmt.Errorf("tx failed: code:%v", res.Code)
	}

	return nil
}
