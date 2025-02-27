package event

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	oracletypes "github.com/youngjoon-lee/dhub/x/oracle/types"
	"github.com/youngjoon-lee/doracle-poc/pkg/dhub/tx"
	"github.com/youngjoon-lee/doracle-poc/pkg/secp256k1"
	"github.com/youngjoon-lee/doracle-poc/pkg/sgx"
)

type JoinEvent struct {
	oraclePrivKey *btcec.PrivateKey
	txExecutor    tx.Executor
}

func NewJoinEvent(oraclePrivKey *btcec.PrivateKey, txExecutor tx.Executor) JoinEvent {
	return JoinEvent{
		oraclePrivKey: oraclePrivKey,
		txExecutor:    txExecutor,
	}
}

func (e JoinEvent) Name() string {
	return "join"
}

func (e JoinEvent) Query() string {
	return "tm.event='Tx' AND message.module='oracle' AND message.action='join'"
}

func (e JoinEvent) Handler(event ctypes.ResultEvent) error {
	log.Debugf("JOIN EVENT: %v", event)

	joinID, err := strconv.ParseUint(event.Events["join.id"][0], 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse join.id: %w", err)
	}

	enclaveReportBase64 := event.Events["join.enclave_report_base64"][0]
	enclaveReport, err := base64.StdEncoding.DecodeString(enclaveReportBase64)
	if err != nil {
		return fmt.Errorf("failed to decode join.enclave_report_base64: %w", err)
	}

	encPubKeyBase64 := event.Events["join.enc_pub_key_base64"][0]
	encPubKeyBytes, err := base64.StdEncoding.DecodeString(encPubKeyBase64)
	if err != nil {
		return fmt.Errorf("failed to decode join.enc_pub_key_base64: %w", err)
	}
	encPubkey, err := secp256k1.PubKeyFromBytes(encPubKeyBytes)
	if err != nil {
		return fmt.Errorf("invalid encryption public key: %w", err)
	}

	voteOption := oracletypes.OptionYes
	encPubKeyHash := sha256.Sum256(encPubKeyBytes)
	if err := sgx.VerifyRemoteReport(enclaveReport, encPubKeyHash[:]); err != nil {
		log.Infof("SGX report verification failed: %v", err)
		voteOption = oracletypes.OptionNo
	}

	yesValue := ""
	if voteOption == oracletypes.OptionYes {
		encryptedOraclePrivKey, err := secp256k1.Encrypt(encPubkey, e.oraclePrivKey.Serialize())
		if err != nil {
			return fmt.Errorf("failed to encrypt oracle priv key: %w", err)
		}
		yesValue = base64.StdEncoding.EncodeToString(encryptedOraclePrivKey)
	}

	if err := e.txExecutor.VoteForJoin(joinID, voteOption, yesValue); err != nil {
		return fmt.Errorf("failed to vote for join: %w", err)
	}

	return nil
}
