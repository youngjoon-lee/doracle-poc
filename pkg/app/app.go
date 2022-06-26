package app

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dhubapp "github.com/youngjoon-lee/dhub/app"
	"github.com/youngjoon-lee/doracle-poc/pkg/dhub/event"
	"github.com/youngjoon-lee/doracle-poc/pkg/dhub/tx"
	"github.com/youngjoon-lee/doracle-poc/pkg/secp256k1"
)

type App struct {
	oraclePrivKey *btcec.PrivateKey
	txExecutor    tx.Executor
	subscriber    *event.Subscriber
}

func NewApp(tendermintRPCAddr, chainID, operatorMnemonic string) (*App, error) {
	setDHubConfig()

	operatorPrivKey, operatorAddr, err := secp256k1.PrivateKeyFromMnemonic(operatorMnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key from mnemonic: %w", err)
	}

	txExecutor, err := tx.NewExecutor(tendermintRPCAddr, chainID, operatorAddr, operatorPrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed to init tx executor: %w", err)
	}

	subscriber, err := event.NewSubscriber(tendermintRPCAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to init subscriber: %w", err)
	}

	if err := subscriber.Start(); err != nil {
		return nil, fmt.Errorf("failed to start subscriber: %w", err)
	}

	return &App{
		oraclePrivKey: nil,
		txExecutor:    txExecutor,
		subscriber:    subscriber,
	}, nil
}

func (app *App) Close() {
	app.subscriber.Stop()
}

func (app *App) SetOraclePrivKey(privKey *btcec.PrivateKey) {
	app.oraclePrivKey = privKey
}

func (app *App) OraclePrivKey() *btcec.PrivateKey {
	return app.oraclePrivKey
}

func (app *App) TxExecutor() tx.Executor {
	return app.txExecutor
}

func (app *App) Subscriber() *event.Subscriber {
	return app.subscriber
}

func (app *App) SubscribeAll() error {
	for _, ev := range app.events() {
		if err := app.Subscriber().Subscribe(ev); err != nil {
			return fmt.Errorf("failed to subscribe: %v: %w", ev.Name(), err)
		}
	}
	return nil
}

func (app *App) events() []event.Event {
	return []event.Event{
		event.NewJoinEvent(app.oraclePrivKey, app.txExecutor),
	}
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
