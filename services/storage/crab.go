package storage

import (
	"github.com/evolutionlandorg/evo-backend/pkg/github.com/regcostajr/go-web3"
	"github.com/evolutionlandorg/evo-backend/pkg/github.com/regcostajr/go-web3/providers"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/roundrobin"
)

type crabCall struct {
	ethCall
}

func (c *crabCall) initClient() {
	provider := providers.NewHTTPProvider(util.GetEnv("CRAB_NODE", "https://crab-rpc.darwinia.network"), 10, true)
	c.RpcClient = roundrobin.New([]interface{}{web3.NewWeb3(provider)})
}
