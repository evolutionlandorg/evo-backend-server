package services

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestDefiBoxSignature(t *testing.T) {
	sig := "sign_version=1.0@access_key=coinwindbj16HX3JDt@secret_key=fevXZwl9bwnc2nRdaS6GGeQH3apxFuJE0Tb3C1XO@timestamp=1616469656521@chain=HECO"
	hash := sha256.New()
	_, _ = hash.Write([]byte(sig))
	bytes := hash.Sum(nil)
	str := base64.StdEncoding.EncodeToString(bytes)
	if str != "BJMv/dmTNW4OvQ1X9m7vfjt3/X+CzazVM3C6rCN/5jg=" {
		t.Errorf("TestDefiBoxSignature failed")
	}
}
