package tx

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ignite-hq/cli/ignite/pkg/cosmoscmd"
	log "github.com/sirupsen/logrus"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"github.com/youngjoon-lee/dhub/app"
)

type Executor struct {
	RpcClient rpcclient.Client
	ChainID   string
	FromAddr  sdk.AccAddress
	PrivKey   cryptotypes.PrivKey
}

func NewExecutor(rpcAddr, chainID string, fromAddr sdk.AccAddress, privKey cryptotypes.PrivKey) (Executor, error) {
	rpcClient, err := client.NewClientFromNode(rpcAddr)
	if err != nil {
		return Executor{}, fmt.Errorf("failed to NewClientFromNode: %w", err)
	}

	return Executor{
		RpcClient: rpcClient,
		ChainID:   chainID,
		FromAddr:  fromAddr,
		PrivKey:   privKey,
	}, nil
}

func (e Executor) Context() client.Context {
	return client.Context{}.WithClient(e.RpcClient).WithChainID(e.ChainID)
}

func (e Executor) signAndBroadcastTx(msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	clientCtx := e.Context()

	encCfg := cosmoscmd.MakeEncodingConfig(app.ModuleBasics)
	txBuilder := encCfg.TxConfig.NewTxBuilder()

	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, fmt.Errorf("failed to set msgs: %w", err)
	}

	//TODO: set fee
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("dhub", sdk.ZeroInt())))

	log.Debugf("retrieving account: %v", e.FromAddr.String())
	accountRetriever := authtypes.AccountRetriever{}
	accNum, accSeq, err := accountRetriever.GetAccountNumberSequence(clientCtx, e.FromAddr)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("failed to get account number/sequence: %w", err)
	}
	log.Debugf("accNum:%v, accSeq:%v", accNum, accSeq)

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	sigV2 := signing.SignatureV2{
		PubKey: e.PrivKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  encCfg.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: accSeq,
	}
	sigsV2 = append(sigsV2, sigV2)

	if err := txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, fmt.Errorf("failed to set signatures: %w", err)
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	signerData := authsigning.SignerData{
		ChainID:       e.ChainID,
		AccountNumber: accNum,
		Sequence:      accSeq,
	}
	sigV2, err = tx.SignWithPrivKey(
		encCfg.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, e.PrivKey, encCfg.TxConfig, accSeq)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with privkey: %w", err)
	}
	sigsV2 = append(sigsV2, sigV2)

	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, fmt.Errorf("failed to set signatures (2nd): %w", err)
	}

	// Generated Protobuf-encoded bytes.
	txBytes, err := encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("failed to encode tx: %w", err)
	}

	log.Debug("broadcasting tx...")
	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast tx: %w", err)
	}

	return res, nil
}
