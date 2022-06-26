package app

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/youngjoon-lee/doracle-poc/pkg/dhub/tx"
)

type App struct {
	oraclePrivKey *btcec.PrivateKey
	txExecutor    tx.Executor
}

func NewApp(oraclePrivKey *btcec.PrivateKey, txExecutor tx.Executor) *App {
	return &App{
		oraclePrivKey: oraclePrivKey,
		txExecutor:    txExecutor,
	}
}

func (app *App) OraclePrivKey() *btcec.PrivateKey {
	return app.oraclePrivKey
}

func (app *App) TxExecutor() tx.Executor {
	return app.txExecutor
}
