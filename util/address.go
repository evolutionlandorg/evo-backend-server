package util

import (
	"encoding/hex"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"
)

/*
ToChecksumAddress converts a string to the proper EIP55 casing.
Transliteration of this code from the EIP55 wiki page:
	function toChecksumAddress (address) {
	  address = address.toLowerCase().replace('0x', '')
	  var hash = createKeccakHash('keccak256').update(address).digest('hex')
	  var ret = '0x'
	  for (var i = 0; i < address.length; i++) {
		if (parseInt(hash[i], 16) >= 8) {
		  ret += address[i].toUpperCase()
		} else {
		  ret += address[i]
		}
	  }
	  return ret
	}
*/
func ToChecksumAddress(address string) string {
	address = strings.Replace(strings.ToLower(address), "0x", "", 1)
	hash := sha3.NewLegacyKeccak256()
	_, _ = hash.Write([]byte(address))
	sum := hash.Sum(nil)
	digest := hex.EncodeToString(sum)

	b := strings.Builder{}
	b.WriteString("0x")

	for i := 0; i < len(address); i++ {
		a := address[i]
		if a > '9' {
			d, _ := strconv.ParseInt(digest[i:i+1], 16, 8)

			if d >= 8 {
				// Upper case it
				a -= 'a' - 'A'
				b.WriteByte(a)
			} else {
				// Keep it lower
				b.WriteByte(a)
			}
		} else {
			// Keep it lower
			b.WriteByte(a)
		}
	}

	return b.String()
}
