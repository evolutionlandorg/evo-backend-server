package util

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"

	"github.com/shopspring/decimal"

	"github.com/spf13/viper"
)

var (
	Evo                  ApplicationConf
	ApostlePictureServer = GetEnv("APOSTLE_PICTURE_SERVER", "http://127.0.0.1:1337")
	ApostlePictureDir    = GetEnv("APOSTLE_PICTURE_DIR", "./images/apostle")

	ApiServerHost = GetEnv("API_SERVER_HOST", "http://127.0.0.1:2333")
)

type ApplicationConf struct {
	ContractsListen       []string
	GenesisApostlePicture map[string]string
	// multi network
	NetworkId     map[string]string // takeBack networkId
	TokenDecimals map[string]int
	Contracts     map[string]ContractAddress
	GraphUrl      map[string]string
	WipeBlock     map[string]WipeBlockConf
	RpcEndpoint   map[string]string
	District      map[string]uint
	ChainId       map[string]uint
	TokenIdPrefix map[string]string
	ApostlePrefix map[string]string
	LandRange     map[string][]int
	GRLandId      map[string][]string
	Formula       map[string][]Formula
	tokens        map[string]map[string]Token
}

type Formula struct {
	Id               int               `json:"id"`
	Index            int               `json:"index"`
	Name             string            `json:"name"`
	Pic              string            `json:"pic"`
	Class            int               `json:"class"` // 阶层
	Grade            int               `json:"grade"` // 等级
	CanDisenchant    bool              `json:"can_disenchant"`
	MajorId          int               `json:"major_id"`
	Minor            FormulaMinorToken `json:"minor"`
	Sort             int               `json:"sort"`
	Issued           int               `json:"issued"`
	Productivity     []decimal.Decimal `json:"productivity"`
	ProtectionPeriod int               `json:"protection_period"`
	ObjectClassExt   int               `json:"objectClassExt"`
}
type FormulaMinorToken struct {
	LP      decimal.Decimal `json:"LP"`
	Element decimal.Decimal `json:"element"`
}

type DrillMinorToken struct {
	Token  string          `json:"token"`
	Amount decimal.Decimal `json:"amount"`
}

type WipeBlockConf struct {
	InitBlock  uint64 `json:"initBlock"`
	CacheKey   string `json:"cacheKey"`
	BlockDelay uint   `json:"blockDelay"`
	BlockTime  uint   `json:"blockTime"`
}

type NetworkConf struct {
	TokenDecimals int             `json:"token_decimals"`
	District      uint            `json:"district"`
	TokenIdPrefix string          `json:"tokenIdPrefix"`
	ApostlePrefix string          `json:"apostlePrefix"`
	GRLand        string          `json:"grLand"`
	LandRange     []int           `json:"range"`
	ChainId       uint            `json:"chainId"`
	Production    *NetworkEnvConf `json:"production"`
	Dev           *NetworkEnvConf `json:"dev"`
	Formula       []Formula       `json:"formula"`
}

type NetworkEnvConf struct {
	Contracts ContractAddress  `json:"contracts"`
	NetworkId string           `json:"networkId"`
	Rpc       string           `json:"rpc"`
	GraphUrl  string           `json:"graphUrl"`
	WipeBlock WipeBlockConf    `json:"wipeBlock"`
	Tokens    map[string]Token `json:"tokens"`
}

type Token struct {
	Address  string `json:"address" example:"0xb52FBE2B925ab79a821b261C82c5Ba0814AAA5e0"`
	Symbol   string `json:"symbol" example:"RING"`
	Decimals int64  `json:"decimals" example:"18"`
}

type ContractAddress map[string]string // map[address]contractName

func LoadConf() {
	var (
		conf ApplicationConf
	)
	path := GetEnv("CONF_DIR", "config")
	if _, err := os.Stat(fmt.Sprintf("%s/application.json", path)); os.IsNotExist(err) {
		path = "../config"
	}
	if _, err := os.Stat(ApostlePictureDir); err != nil {
		Panic(os.MkdirAll(ApostlePictureDir, os.ModePerm), "Fatal error config file")
	}
	viper.SetConfigType("json")
	viper.SetConfigName("application")
	viper.AddConfigPath(path)
	Panic(viper.ReadInConfig(), "Fatal error config file")

	conf.ContractsListen = viper.GetStringSlice("contracts")

	NetworkId := make(map[string]string)
	Contracts := make(map[string]ContractAddress)
	GraphUrl := make(map[string]string)
	TokenDecimals := make(map[string]int)
	WipeBlock := make(map[string]WipeBlockConf)
	RpcEndpoint := make(map[string]string)
	TokenIdPrefix := make(map[string]string)
	ApostlePrefix := make(map[string]string)
	District := make(map[string]uint)
	ChainId := make(map[string]uint)
	LandRange := make(map[string][]int)
	PLOLandId := make(map[string][]string)
	formula := make(map[string][]Formula)
	tokens := make(map[string]map[string]Token)

	conf.GenesisApostlePicture = viper.GetStringMapString("genesisApostlePicture")
	if len(conf.GenesisApostlePicture) == 0 {
		conf.GenesisApostlePicture = make(map[string]string)
	}

	for _, network := range viper.GetStringSlice("network") {
		c, err := os.ReadFile(fmt.Sprintf(path+"/%s.json", strings.ToLower(network)))
		Panic(err)
		var n NetworkConf
		err = json.Unmarshal(c, &n)
		if err != nil {
			panic(err)
		}
		ncf := n.Dev
		if IsProduction() {
			ncf = n.Production
		}
		tokens[network] = make(map[string]Token)
		for key := range ncf.Tokens {
			tokens[network][strings.ToLower(key)] = ncf.Tokens[key]
		}
		if len(tokens[network]) == 0 {
			Panic(errors.New("no tokens"), "Fatal error config file")
		}
		NetworkId[network] = ncf.NetworkId
		contractsMap := make(ContractAddress)
		for name, addr := range ncf.Contracts {
			contractsMap[strings.ToLower(addr)] = name
		}
		TokenDecimals[network] = n.TokenDecimals
		if n.TokenDecimals == 0 {
			TokenDecimals[network] = 18
		}
		Contracts[network] = contractsMap
		GraphUrl[network] = ncf.GraphUrl
		WipeBlock[network] = ncf.WipeBlock
		RpcEndpoint[network] = ncf.Rpc
		ChainId[network] = n.ChainId
		District[network] = n.District
		TokenIdPrefix[network] = n.TokenIdPrefix
		ApostlePrefix[network] = n.ApostlePrefix
		LandRange[network] = n.LandRange
		PLOLandId[network] = strings.Split(n.GRLand, ",")
		formula[network] = n.Formula
	}
	conf.NetworkId = NetworkId
	conf.Contracts = Contracts
	conf.GraphUrl = GraphUrl
	conf.WipeBlock = WipeBlock
	conf.RpcEndpoint = RpcEndpoint
	conf.District = District
	conf.ChainId = ChainId
	conf.TokenIdPrefix = TokenIdPrefix
	conf.ApostlePrefix = ApostlePrefix
	conf.LandRange = LandRange
	conf.GRLandId = PLOLandId
	conf.Formula = formula
	conf.TokenDecimals = TokenDecimals
	conf.tokens = tokens
	Evo = conf
}

func GetNetworkNameById(id string) string {
	for k, v := range Evo.NetworkId {
		if v == id {
			return k
		}
	}
	return ""
}

func (a ApplicationConf) GetToken(chain, token string) *Token {
	if _, ok := a.tokens[chain]; !ok {
		return nil
	}
	var result = new(Token)
	_ = copier.Copy(result, a.tokens[chain][strings.ToLower(token)])
	if result.Address == "" || result.Symbol == "" {
		return nil
	}
	return result
}

func (a ApplicationConf) GetTokens(chain string) (result []Token) {
	for key := range a.tokens[chain] {
		result = append(result, a.tokens[chain][key])
	}
	return result
}
