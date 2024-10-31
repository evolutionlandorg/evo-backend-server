package util

import (
	"strings"

	"github.com/evolutionlandorg/evo-backend/pkg/github.com/regcostajr/go-web3"
	"github.com/evolutionlandorg/evo-backend/pkg/github.com/regcostajr/go-web3/providers"
	"github.com/evolutionlandorg/evo-backend/util/roundrobin"
)

var (
	initEthRPC []interface{}
	Rb         *roundrobin.Balancer
)

func init() {
	rpc := GetEnv("ETH_RPC", "")
	ssl := GetEnv("SSL", "true")

	for _, endpoint := range strings.Split(rpc, ",") {
		EthProvider := providers.NewHTTPProvider(endpoint, 10, ssl == "true")
		initEthRPC = append(initEthRPC, web3.NewWeb3(EthProvider))
	}
	Rb = roundrobin.New(initEthRPC)
}

func GetContractAddress(contract string, chain ...string) (addr string) {
	chainName := "Eth"
	if len(chain) > 0 && chain[0] != "" {
		chainName = chain[0]
	}
	for addr, key := range Evo.Contracts[chainName] {
		if strings.EqualFold(contract, key) {
			return addr
		}
	}
	return ""
}

// GetTokenDecimals 获取 token的精度默认18位
func GetTokenDecimals(chain string) int32 {
	if chain == "" {
		chain = "Eth"
	}
	if decimals, ok := Evo.TokenDecimals[chain]; ok {
		return int32(decimals)
	}
	return 18
}

func GetContractsMap(chain string) (contractMap map[string]string) {
	if chain == "" {
		chain = "Eth"
	}
	contractMap = Evo.Contracts[chain]
	return contractMap
}

func GetContracts(chain string) (contracts []string) {
	contractMap := Evo.Contracts[chain]
	for addr := range contractMap {
		if addr != "" {
			contracts = append(contracts, addr)
		}
	}
	return
}

func LogAnalysis(log string) []string {
	log = TrimHex(log)
	logLength := len(log)
	var logSlice []string
	for i := 0; i < logLength/64; i++ {
		logSlice = append(logSlice, log[i*64:(i+1)*64])
	}
	return logSlice
}
