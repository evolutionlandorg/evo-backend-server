package crypto

import (
	"encoding/hex"
	"reflect"
	"testing"

	"golang.org/x/crypto/ed25519"
)

func TestEd25519SignAndVerify(t *testing.T) {
	kp, err := GenerateEd25519Keypair()
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("helloworld")
	sig, _ := kp.Sign(msg)
	c := make([]byte, hex.EncodedLen(len(sig)))
	hex.Encode(c, sig)
	ok := Ed25519Verify(kp.Public(), msg, sig)
	if !ok {
		t.Fatal("Fail: did not verify ed25519 sig")
	}
}

func TestPublicKeys(t *testing.T) {
	kp, err := GenerateEd25519Keypair()
	if err != nil {
		t.Fatal(err)
	}

	kp2 := NewEd25519Keypair(ed25519.PrivateKey(*(kp.Private())))
	if !reflect.DeepEqual(kp.Public(), kp2.Public()) {
		t.Fatal("Fail: pubkeys do not match")
	}
}

func TestEncodeAndDecodePrivateKey(t *testing.T) {
	kp, err := GenerateEd25519Keypair()
	if err != nil {
		t.Fatal(err)
	}

	enc := kp.Private().Encode()
	res := new(Ed25519PrivateKey)
	err = res.Decode(enc)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(res, kp.Private()) {
		t.Fatalf("Fail: got %x expected %x", res, kp.Private())
	}
}

func TestEncodeAndDecodePublicKey(t *testing.T) {
	kp, err := GenerateEd25519Keypair()
	if err != nil {
		t.Fatal(err)
	}

	enc := kp.Public().Encode()
	res := new(Ed25519PublicKey)
	err = res.Decode(enc)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(res, kp.Public()) {
		t.Fatalf("Fail: got %x expected %x", res, kp.Public())
	}
}

func TestNewEd25519PrivateKey(t *testing.T) {
	//secret := []byte{65, 108, 105, 99, 101, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 209, 114, 167, 76, 218, 76, 134, 89, 18, 195, 43, 160, 168, 10, 87, 174, 105, 171, 174, 65, 14, 92, 203, 89, 222, 32, 78, 47, 68, 50, 219, 79}
	secret := []byte{50, 61, 64, 204, 15, 30, 145, 223, 233, 91, 56, 118, 163, 69, 87, 73, 73, 53, 46, 240, 36, 137, 165, 149, 52, 116, 147, 86, 17, 14, 217, 251, 81, 231, 146, 93, 210, 138, 57, 58, 153, 118, 25, 168, 170, 240, 104, 0, 90, 39, 56, 144, 62, 192, 193, 60, 77, 186, 18, 45, 54, 60, 153, 124}
	//fmt.Println(BytesToHex(secret))
	pk, err := NewEd25519PrivateKey(secret)
	if err != nil || pk == nil {
		t.Fatal(err)
	}
	kp := NewEd25519Keypair(ed25519.PrivateKey(*pk))
	dataBytes := []byte{
		6, 0, 255, 215, 86, 142, 95, 10, 126, 218, 103, 168, 38, 145, 255, 55, 154, 196, 187, 164, 249, 201, 184, 89, 254, 119, 155, 93, 70, 54, 59, 97, 173, 45, 185, 229, 108,
		0,
		4,
		8,
		123,
		0, 0, 0,
		220, 209, 52, 103, 1, 202, 131, 150, 73, 110, 82, 170, 39, 133, 177, 116, 141, 235, 109, 176, 149, 81, 183, 33, 89, 220, 179, 224, 137, 145, 2, 91,
		236, 122, 250, 241, 204, 167, 32, 206, 136, 193, 209, 182, 137, 216, 31, 5, 131, 204, 21, 169, 125, 98, 28, 240, 70, 221, 154, 191, 96, 94, 242, 47,
	}
	sign, _ := kp.Sign(dataBytes)
	c := make([]byte, hex.EncodedLen(len(sign)))
	hex.Encode(c, sign)
	ok := Ed25519Verify(kp.public, dataBytes, sign)
	if !ok {
		t.Fatal("Fail: did not verify ed25519 sig")
	}
}
