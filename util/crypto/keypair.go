package crypto

// KeyType str
type KeyType = string

// Ed25519Type ed25519
const Ed25519Type KeyType = "ed25519"

//Sr25519Type sr25519
const Sr25519Type KeyType = "sr25519"

//Secp256k1Type secp256k1
const Secp256k1Type KeyType = "secp256k1"

type Keypair interface {
	Sign(msg []byte) ([]byte, error)
	Public() PublicKey
	Private() PrivateKey
}

type PublicKey interface {
	Verify(msg, sig []byte) bool
	Encode() []byte
	Decode([]byte) error
}

type PrivateKey interface {
	Sign(msg []byte) ([]byte, error)
	Public() (PublicKey, error)
	Encode() []byte
	Decode([]byte) error
}

//var ss58Prefix = []byte("SS58PRE")

// PublicKeyToAddress returns an ss58 address given a PublicKey
// see: https://github.com/paritytech/substrate/wiki/External-Address-Format-(SS58)
// also see: https://github.com/paritytech/substrate/blob/master/primitives/core/src/crypto.rs#L275
// func PublicKeyToAddress(pub PublicKey) string {
// 	enc := append([]byte{42}, pub.Encode()...)
// 	hasher, err := blake2b.New(64, nil)
// 	if err != nil {
// 		return ""
// 	}
// 	_, err = hasher.Write(append(ss58Prefix, enc...))
// 	if err != nil {
// 		return ""
// 	}
// 	checksum := hasher.Sum(nil)
// 	return base58.Encode(append(enc, checksum[:2]...))
// }

// PublicAddressToByteArray returns []byte address for given PublicKey Address
// func PublicAddressToByteArray(addr string) []byte {
// 	k := base58.Decode(addr)
// 	return k[1:33]
// }

//func DecodePrivateKey(in []byte) (PrivateKey, error) {
//	priv, err := NewEd25519PrivateKey(in)
//	if err != nil {
//		return nil, err
//	}
//
//	return priv, nil
//}
