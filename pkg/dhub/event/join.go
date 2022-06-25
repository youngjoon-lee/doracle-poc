package event

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/youngjoon-lee/doracle-poc/pkg/sgx"
)

type joinEvent struct{}

func (e joinEvent) Name() string {
	return "join"
}

func (e joinEvent) Query() string {
	return "tm.event='Tx' AND message.module='oracle' AND message.action='join'"
}

func (e joinEvent) Handler(event ctypes.ResultEvent) error {
	log.Debugf("JOIN EVENT: %v", event)

	enclaveReportBase64 := event.Events["join.enclave_report_base64"][0]
	enclaveReport, err := base64.StdEncoding.DecodeString(enclaveReportBase64)
	if err != nil {
		return fmt.Errorf("failed to decode join.enclave_report_base64: %w", err)
	}

	encPubKeyBase64 := event.Events["join.enc_pub_key_base64"][0]
	encPubKey, err := base64.StdEncoding.DecodeString(encPubKeyBase64)
	if err != nil {
		return fmt.Errorf("failed to decode join.enc_pub_key_base64: %w", err)
	}

	encPubKeyHash := sha256.Sum256(encPubKey)
	if err := sgx.VerifyRemoteReport(enclaveReport, encPubKeyHash[:]); err != nil {
		return fmt.Errorf("failed to decode join.enc_pub_key_base64: %w", err)
	}

	// TOOD: vote

	return nil
}
