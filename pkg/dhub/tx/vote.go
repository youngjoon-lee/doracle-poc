package tx

import (
	"fmt"

	oracletypes "github.com/youngjoon-lee/dhub/x/oracle/types"
)

func (e Executor) VoteForJoin(joinID uint64, option oracletypes.VoteOption, yesValue string) error {
	msg := oracletypes.NewMsgVoteForJoin(joinID, option, yesValue, e.FromAddr.String())

	res, err := e.signAndBroadcastTx(msg)
	if err != nil {
		return fmt.Errorf("failed to sign and broadcast tx: %w", err)
	}

	if res.Code != 0 {
		return fmt.Errorf("tx failed: code:%v", res.Code)
	}

	return nil
}
