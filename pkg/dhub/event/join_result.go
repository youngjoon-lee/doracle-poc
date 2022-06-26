package event

import (
	"encoding/base64"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	oracletypes "github.com/youngjoon-lee/dhub/x/oracle/types"
	"github.com/youngjoon-lee/doracle-poc/pkg/secp256k1"
	"github.com/youngjoon-lee/doracle-poc/pkg/sgx"
)

type JoinResultEvent struct {
	joinID            uint64
	encPrivKey        *btcec.PrivateKey
	oracleKeyFilePath string
}

func NewJoinResultEvent(joinID uint64, encPrivKey *btcec.PrivateKey, oracleKeyFilePath string) JoinResultEvent {
	return JoinResultEvent{
		joinID: joinID,
		encPrivKey: encPrivKey,
		oracleKeyFilePath: oracleKeyFilePath,
	}
}

func (e JoinResultEvent) Name() string {
	return "join_result"
}

func (e JoinResultEvent) Query() string {
	return fmt.Sprintf(
		"message.module='oracle' AND join_result.id='%v'",
		e.joinID,
	)
}

func (e JoinResultEvent) Handler(event ctypes.ResultEvent) error {
	status := oracletypes.JoinStatus(
		oracletypes.JoinStatus_value[event.Events["join_result.status"][0]],
	)
	if status != oracletypes.JOIN_STATUS_APPROVED {
		return fmt.Errorf("join status not approved: %v", status.String())
	}

	encryptedOraclePrivKey, err := base64.StdEncoding.DecodeString(event.Events["join_result.value"][0])
	if err != nil {
		return fmt.Errorf("failed to decode encryptedOraclePrivKey: %w", err)
	}

	oraclePrivKeyBytes, err := secp256k1.Decrypt(e.encPrivKey, encryptedOraclePrivKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt oraclePrivKeyBytes: %w", err)
	}

	if err := sgx.SealToFile(oraclePrivKeyBytes, e.oracleKeyFilePath); err != nil {
		return fmt.Errorf("failed to save oracle key: %w", err)
	}

	return nil
}
