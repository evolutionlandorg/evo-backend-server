package services

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type TypedData struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SignParam struct {
	HashText string   `json:"hash_text"`
	V        *big.Int `json:"v"`
	R        string   `json:"r"`
	S        string   `json:"s"`
	Amount   string   `json:"amount"`
	Nonce    int64    `json:"nonce"`
	CodeHash string   `json:"code_hash"`
}

func SignData(hash []byte, private string) (string, error) {
	privateKey, err := crypto.HexToECDSA(private)
	if err != nil {
		return "", err
	}
	pre := []byte("\x19EvolutionLand Signed Message:\n32")
	hash = crypto.Keccak256(append(pre[:], hash[:]...))
	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return hexutil.Encode(signature), nil
}

func SignDataForUpdateRole(hash []byte, private string) (string, error) {
	privateKey, err := crypto.HexToECDSA(private)
	if err != nil {
		return "", err
	}
	pre := []byte("\x19EvolutionLand Signed Message For Role Updater:\n32")
	hash = crypto.Keccak256(append(pre[:], hash[:]...))
	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return hexutil.Encode(signature), nil
}

func RecoverPersonalSigned(msg, signed string) string {
	hash := []byte(msg)
	pre := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(msg)))

	hash = crypto.Keccak256(append(pre[:], hash[:]...))

	signBytes, _ := hex.DecodeString(util.TrimHex(signed))
	if signBytes[64] >= 27 {
		signBytes[64] -= 27
	}
	pub, err := crypto.SigToPub(hash, signBytes)
	if pub == nil || err != nil {
		return ""
	}
	address := crypto.PubkeyToAddress(*pub)
	return address.Hex()
}

// func VerifyEIP1271(hash []byte, address, signature string) bool {
// 	const Eip1271MagicValue = "1626ba7e"
// 	hash = crypto.Keccak256(hash)
// 	b, err := ioutil.ReadFile("contract/EIP1271.abi")
// 	if err != nil {
// 		panic(err)
// 	}
// 	contract, err := util.EthRPC().Eth.NewContract(string(b))
// 	if err != nil {
// 		panic(err)
// 	}
// 	transParam := dto.TransactionParameters{To: address, Data: "", From: address}
// 	h, err := contract.Call(&transParam, "isValidSignature", util.BytesToHex(hash), util.TrimHex(signature))
// 	if err != nil || h.Error != nil {
// 		return false
// 	} else {
// 		return util.TrimHex(h.Result.(string))[0:8] == Eip1271MagicValue
// 	}
// }
