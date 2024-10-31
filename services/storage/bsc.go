package storage

import (
	"github.com/evolutionlandorg/evo-backend/pkg/github.com/regcostajr/go-web3"
	"github.com/evolutionlandorg/evo-backend/pkg/github.com/regcostajr/go-web3/providers"
	"github.com/evolutionlandorg/evo-backend/util/roundrobin"
)

type bscCall struct {
	ethCall
}

func (c *bscCall) initClient() {
	c.RpcClient = roundrobin.New([]interface{}{
		web3.NewWeb3(providers.NewHTTPProvider("data-seed-prebsc-1-s1.binance.org:8545", 10, true)),
		web3.NewWeb3(providers.NewHTTPProvider("data-seed-prebsc-2-s1.binance.org:8545", 10, true)),
	})
}
