package storage

import (
	"fmt"
	"github.com/evolutionlandorg/evo-backend/config"
	"os"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("CONF_DIR", "../../config")
	os.Setenv("ABI_PATH", "../../")
	config.InitApplication()
}

func TestCall_ApostleCurrentPriceInToken(t *testing.T) {
	config.InitApplication()
	c := New(Heco)
	assert.True(t, c.LandCurrentPriceInToken("0x2a04000104000101000000000000000400000000000000000000000000000277").GreaterThan(decimal.Zero))
}

func TestCall_ProtectPeriod(t *testing.T) {
	config.InitApplication()
	c := New(Heco)
	fmt.Println(c.ProtectPeriod("0x8303fa5849d4E7B84A44083BD993733c31f63B3B", "2a04000104000105000000000000000400000000000000000000000000000007"))
}

func TestCall_TokenId2Apostle(t *testing.T) {
	config.InitApplication()
	c := New(Heco)
	fmt.Println(c.TokenId2Apostle("2a04000104000102000000000000000400000000000000000000000000000006"))
}
