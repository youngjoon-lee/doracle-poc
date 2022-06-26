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

const (
	gasLimit = 500000
	denom    = "uhub"
)

type Executor struct {
	rpcClient      rpcclient.Client
	chainID        string
	encodingConfig cosmoscmd.EncodingConfig
	signer         sdk.AccAddress
	signerPrivKey  cryptotypes.PrivKey
}

func NewExecutor(rpcAddr, chainID string, signer sdk.AccAddress, signerPrivKey cryptotypes.PrivKey) (Executor, error) {
	rpcClient, err := client.NewClientFromNode(rpcAddr)
	if err != nil {
		return Executor{}, fmt.Errorf("failed to NewClientFromNode: %w", err)
	}

	return Executor{
		rpcClient:      rpcClient,
		chainID:        chainID,
		encodingConfig: cosmoscmd.MakeEncodingConfig(app.ModuleBasics),
		signer:         signer,
		signerPrivKey:  signerPrivKey,
	}, nil
}

func (e Executor) Context() client.Context {
	return client.Context{}.
		WithClient(e.rpcClient).
		WithChainID(e.chainID).
		WithCodec(e.encodingConfig.Marshaler).
		WithInterfaceRegistry(e.encodingConfig.InterfaceRegistry).
		WithTxConfig(e.encodingConfig.TxConfig).
		WithLegacyAmino(e.encodingConfig.Amino).
		WithBroadcastMode("block")
}

func (e Executor) Signer() sdk.AccAddress {
	return e.signer
}

func (e Executor) signAndBroadcastTx(msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	clientCtx := e.Context()

	txBuilder := e.encodingConfig.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, fmt.Errorf("failed to set msgs: %w", err)
	}

	//TODO: set fee
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(denom, sdk.ZeroInt())))
	txBuilder.SetGasLimit(gasLimit)

	log.Debugf("retrieving account: %v", e.signer.String())
	accNum, accSeq, err := authtypes.AccountRetriever{}.GetAccountNumberSequence(clientCtx, e.signer)
	if err != nil {
		return nil, fmt.Errorf("failed to get account number/sequence: %w", err)
	}
	log.Debugf("accNum:%v, accSeq:%v", accNum, accSeq)

	// First round: gather all the signer infos by using the "set empty signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	sigV2 := signing.SignatureV2{
		PubKey: e.signerPrivKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  clientCtx.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: accSeq,
	}
	sigsV2 = append(sigsV2, sigV2)

	if err := txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, fmt.Errorf("failed to set signatures (1st): %w", err)
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	signerData := authsigning.SignerData{
		ChainID:       e.chainID,
		AccountNumber: accNum,
		Sequence:      accSeq,
	}
	sigV2, err = tx.SignWithPrivKey(
		clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, e.signerPrivKey, clientCtx.TxConfig, accSeq,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with privkey: %w", err)
	}
	sigsV2 = append(sigsV2, sigV2)

	if err := txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, fmt.Errorf("failed to set signatures (2nd): %w", err)
	}

	// generated protobuf-encoded bytes
	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
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
