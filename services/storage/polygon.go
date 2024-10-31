package storage

import (
	"strings"

	"github.com/evolutionlandorg/evo-backend/pkg/github.com/regcostajr/go-web3"
	"github.com/evolutionlandorg/evo-backend/pkg/github.com/regcostajr/go-web3/providers"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/roundrobin"
)

type polygonCall struct {
	ethCall
}

func (c *polygonCall) initClient() {
	var endpoint []interface{}
	for _, rpcUrl := range strings.Split(util.GetEnv("POLYGON_RPC", "https://polygon-mumbai.api.onfinality.io/public"), ",") {
		endpoint = append(endpoint, web3.NewWeb3(providers.NewHTTPProvider(rpcUrl, 15, true)))
	}
	c.RpcClient = roundrobin.New(endpoint)
}
