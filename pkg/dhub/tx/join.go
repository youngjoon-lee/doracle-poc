package tx

import (
	"encoding/hex"
	"fmt"

	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	log "github.com/sirupsen/logrus"
	oracletypes "github.com/youngjoon-lee/dhub/x/oracle/types"
)

func (e Executor) Join(operatorAddress string, enclaveReport []byte, encPubKey *secp256k1.PubKey) (uint64, error) {
	msg := oracletypes.NewMsgJoin(operatorAddress, enclaveReport, encPubKey)

	res, err := e.signAndBroadcastTx(msg)
	if err != nil {
		return 0, fmt.Errorf("failed to sign and broadcast tx: %w", err)
	}
	log.Debugf("tx res: code:%v", res.Code)
	if res.Code != 0 {
		return 0, fmt.Errorf("tx failed: code:%v", res.Code)
	}

	dataBytes, err := hex.DecodeString(res.Data)
	if err != nil {
		return 0, fmt.Errorf("failed to decode tx response data: %w", err)
	}

	var response oracletypes.MsgJoinResponse
	if err := e.Context().Codec.Unmarshal(dataBytes, &response); err != nil {
		return 0, fmt.Errorf("failed to unmarshal tx response data: %w", err)
	}

	return response.ID, nil
}
