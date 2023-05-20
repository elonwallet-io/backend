package handlers

import (
	"context"
	"fmt"
	"github.com/Leantar/elonwallet-backend/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"math/big"
)

const (
	MumbaiRPC = "https://rpc-mumbai.maticvigil.com/"
)

func verifyPersonalSignature(message, signature, address string) (bool, error) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(msg))

	sig, err := hexutil.Decode(signature)
	if err != nil {
		return false, err
	}

	// Last byte from personal sign is incremented by 27. We must decrement it to get the recovery id
	// See https://stackoverflow.com/questions/69762108/implementing-ethereum-personal-sign-eip-191-from-go-ethereum-gives-different-s for more info
	sig[len(sig)-1] -= 27

	pkHex := common.HexToAddress(address).Hex()

	pkSig, err := crypto.Ecrecover(hash.Bytes(), sig)
	if err != nil {
		return false, err
	}

	pkSigHex := common.BytesToAddress(crypto.Keccak256(pkSig[1:])[12:]).Hex()

	sigWithoutV := sig[:len(sig)-1]
	valid := crypto.VerifySignature(pkSig, hash.Bytes(), sigWithoutV)

	return valid && pkHex == pkSigHex, nil
}

func createTransaction(client *ethclient.Client, from, to string, ctx context.Context) (*types.Transaction, error) {
	sender := common.HexToAddress(from)
	receiver := common.HexToAddress(to)
	value := new(big.Int).SetInt64(10000000000000000) //0.01 Matic
	chain := new(big.Int).SetInt64(80001)             //Mumbai Test Network

	feeCap, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest fee cap: %w", err)
	}

	nonce, err := client.PendingNonceAt(ctx, sender)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	tipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest tipCap cap: %w", err)
	}

	dynTx := &types.DynamicFeeTx{
		ChainID:   chain,
		Nonce:     nonce,
		GasFeeCap: feeCap,
		GasTipCap: tipCap,
		Gas:       21000,
		To:        &receiver,
		Value:     value,
		Data:      make([]byte, 0),
	}

	tx := types.NewTx(dynTx)
	return tx, nil
}

func signTransaction(tx *types.Transaction, privateKeyHex string) (*types.Transaction, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal().Caller().Err(err).Msg("failed to convert hex to private key")
	}

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(tx.ChainId()), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return signedTx, nil
}

func sendMumbaiTestMatic(to string, cfg config.WalletConfig, ctx context.Context) error {
	client, err := ethclient.DialContext(ctx, MumbaiRPC)
	if err != nil {
		return fmt.Errorf("failed to dial rpc: %w", err)
	}

	tx, err := createTransaction(client, cfg.Address, to, ctx)
	if err != nil {
		return fmt.Errorf("failed to create tx: %w", err)
	}

	signedTx, err := signTransaction(tx, cfg.PrivateKeyHex)
	if err != nil {
		return fmt.Errorf("failed to sign tx: %w", err)
	}

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return fmt.Errorf("failed to send tx: %w", err)
	}

	return nil
}
