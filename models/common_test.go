package models

import (
	"github.com/evolutionlandorg/evo-backend/config"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("CONF_DIR", "../config")
	config.InitApplication()
}

func Test_GetChainByTokenId(t *testing.T) {
	assert.Equal(t, EthChain, GetChainByTokenId("2a010001010001010000000000000001000000000000000000000000000007e9"))
	assert.Equal(t, TronChain, GetChainByTokenId("2a020001020001010000000000000002000000000000000000000000000007e9"))
	assert.Equal(t, CrabChain, GetChainByTokenId("2a03000103000101000000000000000300000000000000000000000000000001"))
}

func Test_getDistrictByChain(t *testing.T) {
	assert.Equal(t, 1, GetDistrictByChain(EthChain))
	assert.Equal(t, 2, GetDistrictByChain(TronChain))
	assert.Equal(t, 3, GetDistrictByChain(CrabChain))
}

func Test_GenerateLandTokenId(t *testing.T) {
	assert.Equal(t, "2a01000101000101000000000000000100000000000000000000000000000001", GenerateLandTokenId(EthChain, 1))
	assert.Equal(t, "2a02000102000101000000000000000200000000000000000000000000000001", GenerateLandTokenId(TronChain, 1))
	assert.Equal(t, "2a03000103000101000000000000000300000000000000000000000000000001", GenerateLandTokenId(CrabChain, 1))
}

func Test_districtByXY(t *testing.T) {
	assert.Equal(t, 1, districtByXY(-85, -10))
	assert.Equal(t, 2, districtByXY(98, 2))
	assert.Equal(t, 3, districtByXY(9, 51))
}

func Test_getAssetTypeByTokenId(t *testing.T) {
	assert.Equal(t, Material, getAssetTypeByTokenId("2a04000b04000b0b000000000000000400000000000000000000000000000000"))
	assert.Equal(t, Material, getAssetTypeByTokenId("2a04000b04000b0b000000000000000400000000000000000000000000000001"))
}
