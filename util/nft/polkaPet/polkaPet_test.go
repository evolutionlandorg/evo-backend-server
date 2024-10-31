package polkaPet

import (
	"context"

	"github.com/evolutionlandorg/evo-backend/config"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/stretchr/testify/assert"

	"os"
	"testing"
)

func init() {
	os.Setenv("CONF_DIR", "../../../config")
	os.Setenv("ABI_PATH", "../../../")
	config.InitApplication()
}

func Test_NftInfo(t *testing.T) {
	pet := New()
	assert.NoError(t, util.InitRedis())
	util.Debug(pet.NftInfo(context.TODO(), "2"))
}

func Test_AllOwnerNft(t *testing.T) {
	pet := &Nft{ContractAddress: "0x3aA012406c56efe330a04F241Fc6477a87C65dc2", TokensIdsCacheKey: Name}
	list, count := pet.AllOwnerNft(context.TODO(), "0xf422673CB7a673f595852f7B00906408A0b073db", nil, 0, 0, "Eth")
	util.Debug(list)
	assert.Greater(t, count, 0)
}
