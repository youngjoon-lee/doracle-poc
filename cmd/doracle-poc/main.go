package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/youngjoon-lee/doracle-poc/cmd/doracle-poc/mode"
	"github.com/youngjoon-lee/doracle-poc/pkg/app"
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

	app, err := app.NewApp(*pTendermintRPC, *pChainID, *pOperatorMnemonic)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}
	defer app.Close()

	if *pInit && *pJoin {
		log.Fatal("do not use -init with -join")
	} else if *pInit {
		if err := mode.Init(app); err != nil {
			log.Fatalf("failed to run the init mode: %v", err)
		}
	} else if *pJoin {
		if err := mode.Join(app); err != nil {
			log.Fatalf("failed to run the join mode: %v", err)
		}
	}

	oraclePrivKeyBytes, err := sgx.UnsealFromFile(mode.OracleKeyFilePath)
	if err != nil {
		log.Fatalf("failed to load and unseal oracle key: %v", err)
	}

	app.SetOraclePrivKey(secp256k1.PrivKeyFromBytes(oraclePrivKeyBytes))
	if err := app.SubscribeAll(); err != nil {
		log.Fatalf("failed to subscribeAll: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh

	log.Info("terminating the process")
}
