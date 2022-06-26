package tx

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	oracletypes "github.com/youngjoon-lee/dhub/x/oracle/types"
)

func (e Executor) VoteForJoin(joinID uint64, option oracletypes.VoteOption, yesValue string) error {
	msg := oracletypes.NewMsgVoteForJoin(joinID, option, yesValue, e.Signer().String())

	res, err := e.signAndBroadcastTx(msg)
	if err != nil {
		return fmt.Errorf("failed to sign and broadcast tx: %w", err)
	}
	log.Debugf("tx res: %v", res)

	if res.Code != 0 {
		return fmt.Errorf("tx failed: code:%v", res.Code)
	}

	return nil
}
