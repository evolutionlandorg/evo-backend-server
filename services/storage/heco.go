package storage

import (
	"github.com/evolutionlandorg/evo-backend/pkg/github.com/regcostajr/go-web3"
	"github.com/evolutionlandorg/evo-backend/pkg/github.com/regcostajr/go-web3/providers"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/roundrobin"
)

type hecoCall struct {
	ethCall
}

func (c *hecoCall) initClient() {
	// MAINNET https://http-mainnet.hecochain.com
	provider := providers.NewHTTPProvider(util.GetEnv("HECO_RPC", "https://http-testnet.hecochain.com"), 10, true)
	c.RpcClient = roundrobin.New([]interface{}{web3.NewWeb3(provider)})
}
