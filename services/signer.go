package services

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/ethereum/go-ethereum/crypto"
	solsha3 "github.com/evolutionlandorg/evo-backend/pkg/github.com/miguelmota/go-solidity-sha3"
)

func AbiEncodingMethod(content string) string {
	hash := solsha3.SoliditySHA3(
		solsha3.String(content),
	)
	return util.AddHex(hex.EncodeToString(hash))
}

func StringToBytes32(content string) []byte {
	hash := solsha3.SoliditySHA3(
		solsha3.String(content),
	)
	return hash
}

func VerifySign(sign, address, challenge string) bool {
	msg := fmt.Sprintf("welcome to evolution land %s", challenge)

	// if personalSigned := strings.EqualFold(address, RecoverPersonalSigned(msg, sign)); personalSigned {
	// 	return true
	// }
	// try eip1271
	return strings.EqualFold(address, RecoverPersonalSigned(msg, sign))
}

func VerifyTronSign(sign, address, challenge string) bool {
	msg := fmt.Sprintf("welcome to evolution land %s", challenge)
	pre := []byte("\x19TRON Signed Message:\n32")
	hash := crypto.Keccak256(append(pre[:], []byte(msg)[:]...))
	signBytes, _ := hex.DecodeString(util.TrimHex(sign))
	return util.TrxVerifySignature(util.TrxHex2Base58Address(address), hash, signBytes)
}

func VerifyTronSignMsg(sign, address, msg string) bool {
	pre := []byte("\x19TRON Signed Message:\n32")
	hash := crypto.Keccak256(append(pre[:], []byte(msg)[:]...))
	signBytes, _ := hex.DecodeString(util.TrimHex(sign))
	return strings.EqualFold(address, util.GetTrxAddressFromSig(hash, signBytes))
}
