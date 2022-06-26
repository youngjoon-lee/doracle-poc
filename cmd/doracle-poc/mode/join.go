package mode

import (
	"context"
	"crypto/sha256"
	"fmt"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	log "github.com/sirupsen/logrus"
	"github.com/youngjoon-lee/doracle-poc/pkg/dhub/event"
	"github.com/youngjoon-lee/doracle-poc/pkg/dhub/tx"
	"github.com/youngjoon-lee/doracle-poc/pkg/secp256k1"
	"github.com/youngjoon-lee/doracle-poc/pkg/sgx"
)

func Join(txExecutor tx.Executor, subscriber *event.Subscriber) error {
	encPrivKey, err := secp256k1.NewPrivKey()
	if err != nil {
		return fmt.Errorf("failed to generate encryption key: %w", err)
	}

	pubKey := &cosmossecp256k1.PubKey{
		Key: encPrivKey.PubKey().SerializeCompressed(),
	}

	pubKeyHash := sha256.Sum256(pubKey.Key)
	enclaveReport, err := sgx.GenerateRemotePeport(pubKeyHash[:])
	if err != nil {
		return fmt.Errorf("failed to generate SGX remote report: %w", err)
	}

	log.Info("SGX report generated. executing tx...")
	joinID, err := txExecutor.Join(txExecutor.FromAddr.String(), enclaveReport, pubKey)
	if err != nil {
		return fmt.Errorf("failed to execute join tx: %w", err)
	}

	log.Info("subscribing the join result...")
	ev := event.NewJoinResultEvent(joinID, encPrivKey, OracleKeyFilePath)
	if err := subscriber.SubscribeOnce(context.Background(), ev); err != nil {
		return fmt.Errorf("failed to subscribe once: %w", err)
	}

	return nil
}
