package mode

import (
	"crypto/sha256"
	"fmt"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	log "github.com/sirupsen/logrus"
	"github.com/youngjoon-lee/doracle-poc/pkg/app"
	"github.com/youngjoon-lee/doracle-poc/pkg/secp256k1"
	"github.com/youngjoon-lee/doracle-poc/pkg/sgx"
)

const OracleKeyFilePath = "/data/oracle-key.sealed"

func Init(app *app.App) error {
	oraclePrivKey, err := secp256k1.NewPrivKey()
	if err != nil {
		log.Fatalf("failed to generate oracle key: %v", err)
	}

	if err := sgx.SealToFile(oraclePrivKey.Serialize(), OracleKeyFilePath); err != nil {
		log.Fatalf("failed to save oracle key: %v", err)
	}

	oraclePubKey := &cosmossecp256k1.PubKey{
		Key: oraclePrivKey.PubKey().SerializeCompressed(),
	}

	pubKeyHash := sha256.Sum256(oraclePubKey.Key)
	enclaveReport, err := sgx.GenerateRemotePeport(pubKeyHash[:])
	if err != nil {
		return fmt.Errorf("failed to generate SGX remote report: %w", err)
	}

	log.Info("oracle key and SGX report generated. executing tx...")
	txExecutor := app.TxExecutor()
	err = txExecutor.Init(txExecutor.Signer().String(), enclaveReport, oraclePubKey)
	if err != nil {
		return fmt.Errorf("failed to execute join tx: %w", err)
	}

	return nil
}
