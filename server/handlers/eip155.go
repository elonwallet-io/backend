package handlers

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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
