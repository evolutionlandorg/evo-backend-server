package models

import (
	"math/big"
	"sync"

	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/orcaman/concurrent-map"
)

type Currency struct {
	Code            string `json:"code" yaml:"code"`
	ContractAddress string `json:"contract_address" yaml:"contract_address"`
	Chain           string `json:"chain"`
}

func (c Currency) GetBalance(addr string) *big.Int {
	sg := storage.New(c.Chain)
	if c.ContractAddress == "" {
		return sg.GetBalance(addr)
	}
	return sg.BalanceOf(addr, c.ContractAddress)
}

func newCurrency(code, chain string) Currency {
	contract := ""
	if code != chain {
		contract = util.GetContractAddress(code, chain)
	}
	return Currency{Code: code, ContractAddress: contract, Chain: chain}
}

func CurrenciesBalance(currencies []string, address string, chain string) map[string]interface{} {
	balanceMap := cmap.New()
	var wg sync.WaitGroup
	for _, code := range currencies {
		currency := newCurrency(code, chain)
		wg.Add(1)
		go func(code string, currency Currency) {
			defer wg.Done()
			balance := currency.GetBalance(address).String()
			balanceMap.Set(currency.Code, balance)
		}(code, currency)
	}
	wg.Wait()
	return balanceMap.Items()
}
