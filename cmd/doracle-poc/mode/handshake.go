package mode

import (
	log "github.com/sirupsen/logrus"
	"github.com/youngjoon-lee/doracle-poc/pkg/client"
	"github.com/youngjoon-lee/doracle-poc/pkg/secp256k1"
	"github.com/youngjoon-lee/doracle-poc/pkg/sgx"
)

func Handshake(peerAddr string) error {
	nodePrivKey, err := secp256k1.NewPrivKey()
	if err != nil {
		log.Fatalf("failed to generate node key: %v", err)
	}

	oraclePrivKey, err := client.CallHandshake(peerAddr, nodePrivKey)
	if err != nil {
		log.Fatalf("failed to handshake: %v", err)
	}

	if err := sgx.SealToFile(oraclePrivKey.Serialize(), OracleKeyFilePath); err != nil {
		log.Fatalf("failed to save oracle key: %v", err)
	}

	return nil
}
