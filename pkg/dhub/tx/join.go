package tx

import (
	"fmt"
	"strconv"

	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	log "github.com/sirupsen/logrus"
	oracletypes "github.com/youngjoon-lee/dhub/x/oracle/types"
)

func (e Executor) Join(operatorAddress string, enclaveReport []byte, encPubKey *secp256k1.PubKey) (uint64, error) {
	msg := oracletypes.NewMsgJoin(operatorAddress, enclaveReport, encPubKey)

	res, err := e.signAndBroadcastTx(msg)
	if err != nil {
		return 0, fmt.Errorf("failed to sign and broadcast tx: %w", err)
	}
	log.Debugf("tx res:%v", res)
	if res.Code != 0 {
		return 0, fmt.Errorf("tx failed: code:%v", res.Code)
	}

	id, err := getJoinID(res)
	if err != nil {
		return 0, fmt.Errorf("failed to get joinID: %w", err)
	}
	log.Debugf("joinID:%v", id)

	return id, nil
}

func getJoinID(res *sdk.TxResponse) (uint64, error) {
	for _, event := range res.Events {
		if event.Type == oracletypes.EventTypeJoin {
			for _, attr := range event.Attributes {
				if string(attr.Key) == oracletypes.AttributeKeyID {
					id, err := strconv.ParseUint(string(attr.Value), 10, 64)
					if err != nil {
						return 0, fmt.Errorf("failed to parse id: %w", err)
					}
					return id, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("joinID not found from events")
}
