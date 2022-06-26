package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	sdk "github.com/cosmos/cosmos-sdk/types"
	log "github.com/sirupsen/logrus"
	dhubapp "github.com/youngjoon-lee/dhub/app"
	"github.com/youngjoon-lee/doracle-poc/cmd/doracle-poc/mode"
	"github.com/youngjoon-lee/doracle-poc/pkg/app"
	"github.com/youngjoon-lee/doracle-poc/pkg/dhub/event"
	"github.com/youngjoon-lee/doracle-poc/pkg/dhub/tx"
	"github.com/youngjoon-lee/doracle-poc/pkg/secp256k1"
	"github.com/youngjoon-lee/doracle-poc/pkg/sgx"
)

func main() {
	pTendermintRPC := flag.String("tm-rpc", "tcp://127.0.0.1:26657", "tendermint rpc addr")
	pChainID := flag.String("chain-id", "dhub-1", "chain ID")
	pOperatorMnemonic := flag.String("operator", "", "operator mnemonic")
	pInit := flag.Bool("init", false, "run doracle with the init mode")
	pJoin := flag.Bool("join", false, "run doracle with the join mode")
	pDebug := flag.Bool("debug", false, "enable debug logs")
	flag.Parse()

	if *pDebug {
		log.SetLevel(log.DebugLevel)
	}

	setDHubConfig()

	operatorPrivKey, operatorAddr, err := secp256k1.PrivateKeyFromMnemonic(*pOperatorMnemonic)
	if err != nil {
		log.Fatalf("failed to get private key from mnemonic: %w", err)
	}

	txExecutor, err := tx.NewExecutor(*pTendermintRPC, *pChainID, operatorAddr, operatorPrivKey)
	if err != nil {
		log.Fatalf("failed to init tx executor: %v", err)
	}

	subscriber, err := event.NewSubscriber(*pTendermintRPC)
	if err != nil {
		log.Fatalf("failed to init subscriber: %v", err)
	}
	if err := subscriber.Start(); err != nil {
		log.Fatalf("failed to start subscriber: %v", err)
	}
	defer subscriber.Stop()

	if *pInit && *pJoin {
		log.Fatal("do not use -init with -join")
	} else if *pInit {
		if err := mode.Init(); err != nil {
			log.Fatalf("failed to run the init mode: %v", err)
		}
	} else if *pJoin {
		if err := mode.Join(txExecutor, subscriber); err != nil {
			log.Fatalf("failed to run the join mode: %v", err)
		}
	}

	oraclePrivKeyBytes, err := sgx.UnsealFromFile(mode.OracleKeyFilePath)
	if err != nil {
		log.Fatalf("failed to load and unseal oracle key: %v", err)
	}

	app := app.NewApp(secp256k1.PrivKeyFromBytes(oraclePrivKeyBytes), txExecutor)

	if err := subscriber.SubscribeAll(app); err != nil {
		log.Fatalf("failed to subscribeAll: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh

	log.Info("terminating the process")
}

func setDHubConfig() {
	accountAddressPrefix := dhubapp.AccountAddressPrefix

	// Set prefixes
	accountPubKeyPrefix := accountAddressPrefix + "pub"
	validatorAddressPrefix := accountAddressPrefix + "valoper"
	validatorPubKeyPrefix := accountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := accountAddressPrefix + "valcons"
	consNodePubKeyPrefix := accountAddressPrefix + "valconspub"

	// Set and seal config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(accountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	config.Seal()
}
