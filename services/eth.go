package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/spf13/cast"
)

// method id sha3[0:10] == Keccak

type Receipts struct {
	BlockNumber      string `json:"block_number"`
	Logs             []Log  `json:"logs"`
	Status           string `json:"status"`
	ChainSource      string `json:"chainSource"`
	GasUsed          string `json:"gasUsed"`
	LogsBloom        string `json:"logsBloom"`
	Solidity         bool   `json:"solidity"`
	TransactionIndex string `json:"transactionIndex"`
	BlockHash        string `json:"blockHash"`
}

type Log struct {
	Topics  []string `json:"topics"`
	Data    string   `json:"data"`
	Address string   `json:"address"`
}

type GasNow struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  struct {
		SafeGasPrice    string `json:"SafeGasPrice"`
		ProposeGasPrice string `json:"ProposeGasPrice"`
		FastGasPrice    string `json:"FastGasPrice"`
	} `json:"result"`
}

var (
	// defaultValue perror.qin@itering.com https://etherscan.io/myapikey
	apiKey = util.GetEnv("ETH-GAS-KEY", "SJ71ZCRJUVEX2ES7XEHW3GR1RT43QD21S3")
	apiUrl = fmt.Sprintf("https://api.etherscan.io/api?module=gastracker&action=gasoracle&apikey=%s", apiKey)
)

func GetEthGasPrice(ctx context.Context) map[string]interface{} {
	var cache []byte
	if cache = util.GetCache(ctx, "eth:gasnow"); cache == nil {
		res := util.HttpGetWithTimeout(apiUrl, time.Second*10)
		if len(res) == 0 {
			return nil
		}
		var gn GasNow
		err := json.Unmarshal(res, &gn)
		if err != nil {
			return nil
		}
		c := map[string]interface{}{"safe": cast.ToInt(gn.Result.SafeGasPrice) * 1000000000,
			"standard": cast.ToInt(gn.Result.ProposeGasPrice) * 1000000000,
			"fast":     cast.ToInt(gn.Result.SafeGasPrice) * 1000000000}
		_ = util.SetCache(ctx, "eth:gasnow", []byte(util.ToString(c)), 120)
		return c
	} else {
		var c map[string]interface{}
		_ = json.Unmarshal(cache, &c)
		return c
	}
}
