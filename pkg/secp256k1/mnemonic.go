package secp256k1

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/go-bip39"
	dhubapp "github.com/youngjoon-lee/dhub/app"
)

const (
	defaultAccount      = 0
	defaultAddressIndex = 0
	coinType            = uint32(118)
)

func PrivateKeyFromMnemonic(mnemonic string) (cryptotypes.PrivKey, sdk.AccAddress, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, nil, fmt.Errorf("invalid mnemonic")
	}

	hdPath := hd.NewFundraiserParams(defaultAccount, coinType, defaultAddressIndex).String()
	master, ch := hd.ComputeMastersFromSeed(bip39.NewSeed(mnemonic, ""))

	privKeyBytes, err := hd.DerivePrivateKeyForPath(master, ch, hdPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive private key from mnemonic: %w", err)
	}

	bz := make([]byte, 32)
	copy(bz, privKeyBytes)
	privKey := &secp256k1.PrivKey{Key: bz}

	accAddrStr, err := bech32.ConvertAndEncode(dhubapp.AccountAddressPrefix, privKey.PubKey().Address().Bytes())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get account address string: %w", err)
	}
	accAddr, err := sdk.AccAddressFromBech32(accAddrStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse account address string: %w", err)
	}

	return privKey, accAddr, nil
}
