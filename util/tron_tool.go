package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"

	"errors"
)

const (
	// alphabet is the modified base58 alphabet used by Bitcoin.
	alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	alphabetIdx0 = '1'
)

var b58 = [256]byte{
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 0, 1, 2, 3, 4, 5, 6,
	7, 8, 255, 255, 255, 255, 255, 255,
	255, 9, 10, 11, 12, 13, 14, 15,
	16, 255, 17, 18, 19, 20, 21, 255,
	22, 23, 24, 25, 26, 27, 28, 29,
	30, 31, 32, 255, 255, 255, 255, 255,
	255, 33, 34, 35, 36, 37, 38, 39,
	40, 41, 42, 43, 255, 44, 45, 46,
	47, 48, 49, 50, 51, 52, 53, 54,
	55, 56, 57, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
}
var bigRadix = big.NewInt(58)
var bigZero = big.NewInt(0)

// Decode decodes a modified base58 string to a byte slice.
func Base58Decode(b string) []byte {
	answer := big.NewInt(0)
	j := big.NewInt(1)

	scratch := new(big.Int)
	for i := len(b) - 1; i >= 0; i-- {
		tmp := b58[b[i]]
		if tmp == 255 {
			return []byte("")
		}
		scratch.SetInt64(int64(tmp))
		scratch.Mul(j, scratch)
		answer.Add(answer, scratch)
		j.Mul(j, bigRadix)
	}

	tmpval := answer.Bytes()

	var numZeros int
	for numZeros = 0; numZeros < len(b); numZeros++ {
		if b[numZeros] != alphabetIdx0 {
			break
		}
	}
	flen := numZeros + len(tmpval)
	val := make([]byte, flen)
	copy(val[numZeros:], tmpval)

	return val
}

// Base58Encode encodes a byte slice to a modified base58 string.
func Base58Encode(b []byte) string {
	x := new(big.Int)
	x.SetBytes(b)

	answer := make([]byte, 0, len(b)*136/100)
	for x.Cmp(bigZero) > 0 {
		mod := new(big.Int)
		x.DivMod(x, bigRadix, mod)
		answer = append(answer, alphabet[mod.Int64()])
	}

	// leading zero bytes
	for _, i := range b {
		if i != 0 {
			break
		}
		answer = append(answer, alphabetIdx0)
	}

	// reverse
	alen := len(answer)
	for i := 0; i < alen/2; i++ {
		answer[i], answer[alen-1-i] = answer[alen-1-i], answer[i]
	}

	return string(answer)
}

// ErrChecksum indicates that the checksum of a check-encoded string does not verify against
// the checksum.
var Base58ErrChecksum = errors.New("checksum error")

// ErrInvalidFormat indicates that the check-encoded string has an invalid format.
var Base58ErrInvalidFormat = errors.New("invalid format: version and/or checksum bytes missing")

// checksum: first four bytes of sha256^2
func checksum(input []byte) (cksum [4]byte) {
	h := sha256.Sum256(input)
	h2 := sha256.Sum256(h[:])
	copy(cksum[:], h2[:4])
	return
}

// Base58CheckEncode prepends a version byte and appends a four byte checksum.
func Base58CheckEncode(input []byte, version byte) string {
	b := make([]byte, 0, 1+len(input)+4)
	b = append(b, version)
	b = append(b, input[:]...)
	cksum := checksum(b)
	b = append(b, cksum[:]...)
	return Base58Encode(b)
}

// Base58CheckDecode decodes a string that was encoded with CheckEncode and verifies the checksum.
func Base58CheckDecode(input string) (result []byte, version byte, err error) {
	decoded := Base58Decode(input)
	if len(decoded) < 5 {
		return nil, 0, Base58ErrInvalidFormat
	}
	version = decoded[0]
	var cksum [4]byte
	copy(cksum[:], decoded[len(decoded)-4:])
	if checksum(decoded[:len(decoded)-4]) != cksum {
		return nil, 0, Base58ErrChecksum
	}
	payload := decoded[1 : len(decoded)-4]
	result = append(result, payload...)
	return
}

//addresss in base58 format
func TrxVerifySignature(address string, hash, signature []byte) bool {
	return address == GetTrxAddressFromSig(hash, signature)
}

func GetTrxAddressFromSig(hash, signature []byte) string {
	if len(signature) < 65 {
		return ""
	}
	signature[64] -= 27
	pub, err := crypto.SigToPub(hash, signature)

	if err != nil || pub == nil {
		return ""
	}
	eaddr := [20]byte(crypto.PubkeyToAddress(*pub))
	last20b := eaddr[:]

	return Base58CheckEncode(last20b, 0x41)
}

func TrxBase58toHexAddress(base58 string) string {

	last20b, version, err := Base58CheckDecode(base58)

	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	buf.WriteByte(version)
	buf.Write(last20b)

	return fmt.Sprintf("%x", buf.Bytes())
}

func TrxHex2Base58Address(hexAddr string) string {

	if !strings.HasPrefix(hexAddr, "41") || len(hexAddr) != 42 {
		return ""
	}

	last20b, _ := hex.DecodeString(hexAddr[2:])

	return Base58CheckEncode(last20b, 0x41)

}
