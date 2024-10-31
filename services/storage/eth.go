package storage

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/evolutionlandorg/block-scan/services"
	"github.com/evolutionlandorg/evo-backend/pkg/github.com/regcostajr/go-web3/complex/types"
	"github.com/evolutionlandorg/evo-backend/pkg/github.com/regcostajr/go-web3/dto"
	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

type ethCall struct {
	Call
}

func NewEth() *ethCall {
	c := Call{Network: "Eth"}
	c.RpcClient = util.Rb
	return &ethCall{Call: c}
}

type BlockResult struct {
	Difficulty      string `json:"difficulty"`
	ExtraData       string `json:"extraData"`
	GasLimit        string `json:"gasLimit"`
	GasUsed         string `json:"gasUsed"`
	Hash            string `json:"hash"`
	LogsBloom       string `json:"logs_bloom"`
	Miner           string `json:"miner"`
	MixHash         string `json:"mixHash"`
	Nonce           string `json:"nonce"`
	Number          string `json:"number"`
	ParentHash      string `json:"parentHash"`
	ReceiptsRoot    string `json:"receiptsRoot"`
	Sha3Uncles      string `json:"sha3Uncles"`
	StateRoot       string `json:"stateRoot"`
	Timestamp       string `json:"timestamp"`
	TotalDifficulty string `json:"totalDifficulty"`
	Transactions    []struct {
		Hash  string `json:"hash"`
		From  string `json:"from"`
		Input string `json:"input"`
		To    string `json:"to"`
		Value string `json:"value"`
	} `json:"transactions"`

	TransactionsRoot string `json:"transactionsRoot"`
}

func (c ethCall) GetTokenLocationHM(tokenId string) (int64, int64, error) {
	landContract := util.GetContractAddress("tokenLocation", c.Network)
	b, err := os.ReadFile("contract/tokenLocation.abi")
	util.Panic(err)
	contract, err := c.EthRPC().Eth.NewContract(string(b))
	if err != nil {
		return 0, 0, err
	}
	transParam := dto.TransactionParameters{To: landContract, Data: ""}
	if h, err := contract.Call(&transParam, "getTokenLocationHM", util.U256(tokenId)); err != nil || h.Error != nil {
		return 0, 0, err
	} else {
		slice := util.LogAnalysis(h.Result.(string))
		x := util.EncodeU256(slice[0]).Int64()
		y := util.EncodeU256(slice[1]).Int64()
		return x, y, nil
	}
}

func (c ethCall) OwnerOf(tokenId string) (string, error) {
	landContract := util.GetContractAddress("objectOwnership", c.Network)
	b, err := os.ReadFile("contract/objectOwnership.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: landContract, Data: ""}
	if h, err := contract.Call(&transParam, "ownerOf", util.U256(tokenId)); err != nil || h.Error != nil {
		return "", err
	} else {
		if len(h.Result.(string)) < 66 {
			return "", errors.New("not address")
		}
		owner := h.Result.(string)
		return util.AddHex(util.TrimHex(owner)[24:64]), nil
	}
}

func (c ethCall) GetResourceRateAttr(tokenId string) (string, error) {
	LandDataContract := util.GetContractAddress("land", c.Network)
	b, err := os.ReadFile("contract/land.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: LandDataContract, Data: ""}
	if h, err := contract.Call(&transParam, "getResourceRateAttr", util.U256(tokenId)); err != nil || h.Error != nil {
		return "", errors.New("not found resource")
	} else {
		return h.Result.(string), nil
	}
}

func (c ethCall) DEGOTokensOfOwner(owner string) []string {
	degoContract := util.GetContractAddress("dego", c.Network)
	b, err := os.ReadFile("contract/dego.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: degoContract, Data: ""}
	if h, err := contract.Call(&transParam, "tokensOfOwner", owner); err != nil || h.Error != nil {
		return nil
	} else {
		if raws := util.LogAnalysis(h.Result.(string)); len(raws) > 2 {
			return raws[2:]
		}
		return nil
	}
}

func (c ethCall) DEGOOwnerOf(tokenId string) string {
	degoContract := util.GetContractAddress("dego", c.Network)
	b, err := os.ReadFile("contract/dego.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: degoContract, Data: ""}
	if h, err := contract.Call(&transParam, "ownerOf", util.U256(tokenId)); err != nil || h.Error != nil {
		return ""
	} else {
		if len(h.Result.(string)) < 66 {
			return ""
		}
		owner := h.Result.(string)
		return util.AddHex(util.TrimHex(owner)[24:64])
	}
}

func (c ethCall) ReceiptLog(tx string) (*services.Receipts, error) {
	params := make([]interface{}, 1)
	params[0] = tx

	var receipt services.Receipts
	pointer := &dto.RequestResult{}
	err := c.EthRPC().Provider.SendRequest(pointer, "eth_getTransactionReceipt", params)
	if err != nil {
		return nil, err
	}

	err = mapstructure.Decode(pointer.Result, &receipt)
	receipt.ChainSource = c.Network
	receipt.Solidity = true
	return &receipt, err
}

func (c ethCall) BlockNumber() uint64 {
	blockNum, err := c.EthRPC().Eth.GetBlockNumber()
	if err != nil {
		return 0
	}
	return blockNum.Uint64()
}

func (c ethCall) FilterTrans(blockNum uint64, filter []string) (txn []string, contracts []string, timestamp uint64, transactionTo []string) {
	block, err := c.BlockInfo(blockNum)
	if err != nil {
		return
	}

	for _, transaction := range block.Transactions {
		to := strings.ToLower(transaction.To)
		transactionTo = append(transactionTo, to)
		if util.StringInSlice(to, filter) {
			txn = append(txn, transaction.Hash)
			contracts = append(contracts, to)
		}
	}
	timestamp = util.U256(block.Timestamp).Uint64()
	return
}

func (c ethCall) BlockInfo(blockNum uint64) (*BlockResult, error) {
	params := []interface{}{util.IntToHex(blockNum), true}
	pointer := &dto.RequestResult{}
	if err := c.EthRPC().Provider.SendRequest(pointer, "eth_getBlockByNumber", params); err != nil {
		return nil, err
	}
	var block BlockResult
	err := mapstructure.Decode(pointer.Result, &block)
	return &block, err
}

func (c ethCall) GetBalance(address string) *big.Int {
	if balance, err := c.EthRPC().Eth.GetBalance(address, "latest"); err == nil {
		return balance
	}
	return big.NewInt(0)
}

func (c ethCall) BalanceOf(address, contractAddress string) *big.Int {
	b, err := os.ReadFile("contract/erc20.abi") // erc20
	util.Panic(err)
	contract, err := c.EthRPC().Eth.NewContract(string(b))
	if err != nil {
		return big.NewInt(0)
	}
	transParam := dto.TransactionParameters{To: contractAddress, Data: ""}
	h, err := contract.Call(&transParam, "balanceOf", address)
	if err != nil || h.Error != nil {
		return big.NewInt(-1)
	}
	resultString := h.Result.(string)
	return util.U256(resultString)
}

func (c ethCall) UserToNonce(address, contractAddress string) int64 {
	b, _ := os.ReadFile("contract/takeBack.abi")
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: contractAddress, Data: ""}
	if h, err := contract.Call(&transParam, "userToNonce", address); err == nil && h.Error == nil {
		return types.ComplexIntResponse(h.Result.(string)).ToInt64()
	}
	return 0
}

func (c ethCall) QuerySwapFee(amount *big.Int) decimal.Decimal {
	contractAddress := util.GetContractAddress("tokenSwap", c.Network)
	b, err := os.ReadFile("contract/tokenSwap.abi")
	util.Panic(err)
	contract, err := c.EthRPC().Eth.NewContract(string(b))
	if err != nil {
		return decimal.Zero
	}
	transParam := dto.TransactionParameters{To: contractAddress, Data: ""}
	if h, err := contract.Call(&transParam, "querySwapFee", amount); err != nil || h.Error != nil {
		return decimal.Zero
	} else {
		slice := util.LogAnalysis(h.Result.(string))
		if len(slice) < 1 {
			return decimal.New(100, 0).Mul(decimal.NewFromFloat(1e18))
		}
		return decimal.NewFromBigInt(util.U256(slice[0]), 0)
	}

}

func (c ethCall) PointsBalanceOf(address string) decimal.Decimal {
	contractAddress := util.GetContractAddress("userPoints", c.Network)
	b, err := os.ReadFile("contract/userPoints.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: contractAddress, Data: ""}
	if h, err := contract.Call(&transParam, "pointsBalanceOf", address); err != nil || h.Error != nil {
		return decimal.Zero
	} else {
		price := util.BigToDecimal(util.U256(h.Result.(string)), util.GetTokenDecimals(c.Network))
		return price
	}
}

func (c ethCall) TotalRewardInPool() decimal.Decimal {
	contractAddress := util.GetContractAddress("pointsReward", c.Network)
	b, err := os.ReadFile("contract/pointsReward.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: contractAddress, Data: ""}
	if h, err := contract.Call(&transParam, "totalRewardInPool", util.GetContractAddress("ring", c.Network)); err != nil || h.Error != nil {
		return decimal.Zero
	} else {
		price := util.BigToDecimal(util.U256(h.Result.(string)), util.GetTokenDecimals(c.Network))
		return price.Round(5)
	}
}

func (c ethCall) ComputePenalty(depositId int64) string {
	bankContract := util.GetContractAddress("bank", c.Network)
	b, err := os.ReadFile("contract/bank.abi")
	util.Panic(err)
	contract, err := c.EthRPC().Eth.NewContract(string(b))
	if err != nil {
		return ""
	}
	transParam := dto.TransactionParameters{To: bankContract, Data: ""}
	if h, err := contract.Call(&transParam, "computePenalty", big.NewInt(depositId)); err != nil {
		return ""
	} else {
		return util.U256(h.Result.(string)).String()
	}
}

func (c ethCall) BlockHeader(blockNum uint64) *services.BlockHeader {
	raw, err := c.BlockInfo(blockNum)
	if err != nil || raw.Hash == "" {
		return nil
	}
	header := services.BlockHeader{BlockTimeStamp: util.U256(raw.Timestamp).Uint64(), Hash: raw.Hash}
	return &header
}

func (c ethCall) GetTransactionStatus(tx string) string {
	res, err := c.ReceiptLog(tx)
	if res == nil || err != nil {
		return ""
	}
	if res.Status == "0x1" {
		return "Success"
	} else if res.Status == "0x0" {
		return "Fail"
	}
	return ""
}

func (c ethCall) GetTransaction(tx string) *Transaction {
	res, _ := c.EthRPC().Eth.GetTransactionByHash(tx)
	if res == nil {
		return nil
	}
	txn := Transaction{BlockNum: res.BlockNumber.Uint64()}
	return &txn
}

func (c ethCall) ApostleCurrentPriceInToken(tokenId string) decimal.Decimal {
	landContract := util.GetContractAddress("clockAuctionApostle", c.Network)
	b, err := os.ReadFile("contract/clockAuctionApostle.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: landContract, Data: ""}
	if h, err := contract.Call(&transParam, "getCurrentPriceInToken", util.U256(tokenId)); err != nil || h.Error != nil {
		return decimal.RequireFromString("-1")
	} else {
		return util.BigToDecimal(util.U256(h.Result.(string)), util.GetTokenDecimals(c.Network))
	}
}

func (c ethCall) LandCurrentPriceInToken(tokenId string) decimal.Decimal {
	landContract := util.GetContractAddress("clockAuction", c.Network)
	dir := util.GetEnv("CONTRACT_ABI", "contract")
	b, err := os.ReadFile(filepath.Join(abiPath(), dir, "clockAuctionApostle.abi"))
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: landContract, Data: ""}
	if h, err := contract.Call(&transParam, "getCurrentPriceInToken", util.U256(tokenId)); err != nil || h.Error != nil {
		return decimal.RequireFromString("-1")
	} else {
		return util.BigToDecimal(util.U256(h.Result.(string)), util.GetTokenDecimals(c.Network))
	}
}

func (c ethCall) LandMask(tokenId string) (int, error) {
	LandDataContract := util.GetContractAddress("land", c.Network)
	b, err := os.ReadFile("contract/land.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: LandDataContract, Data: ""}
	if h, err := contract.Call(&transParam, "getFlagMask", util.U256(tokenId)); err != nil || h.Error != nil {
		return 0, err
	} else {
		data := h.Result.(string)
		return util.StringToInt(util.U256(data).String()), nil
	}
}

func (c ethCall) UnclaimedResource(tokenId string, resourceAddress []string) []string {
	landContract := util.GetContractAddress("landResource", c.Network)
	b, err := os.ReadFile("contract/landResource.abi")
	util.Panic(err)
	contract, err := c.EthRPC().Eth.NewContract(string(b))
	if err != nil {
		return nil
	}
	transParam := dto.TransactionParameters{To: landContract, Data: ""}
	if h, err := contract.Call(&transParam, "availableLandResources", util.U256(tokenId), resourceAddress); err != nil || h.Error != nil {
		return nil
	} else {
		dataSlice := util.LogAnalysis(h.Result.(string))
		if len(dataSlice) <= 2 {
			return nil
		}
		return dataSlice[2:]
	}
}

func (c ethCall) DrillUnclaimedResource(drillTokenId string, resourceAddress []string) []string {
	landContract := util.GetContractAddress("landResource", c.Network)
	b, err := os.ReadFile("contract/landResource.abi")
	util.Panic(err)
	contract, err := c.EthRPC().Eth.NewContract(string(b))
	if err != nil {
		return nil
	}
	transParam := dto.TransactionParameters{To: landContract, Data: ""}
	objectOwnershipAddress := util.GetContractAddress("objectOwnership", c.Network)
	if h, err := contract.Call(&transParam, "availableItemResources", objectOwnershipAddress, util.U256(drillTokenId),
		resourceAddress); err != nil || h.Error != nil {
		return nil
	} else {
		dataSlice := util.LogAnalysis(h.Result.(string))
		if len(dataSlice) <= 2 {
			return nil
		}
		return dataSlice[2:]
	}
}

func (c ethCall) Auction(tokenId string) (string, error) {
	landContract := util.GetContractAddress("clockAuction", c.Network)
	b, err := os.ReadFile("contract/clockAuction.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: landContract}
	h, err := contract.Call(&transParam, "getAuction", util.U256(tokenId))
	if err != nil {
		return "", fmt.Errorf("get auction fail, err: %v", err)
	}
	if h.Error != nil {
		return "", fmt.Errorf("get auction fail, err: %+v", h.Error)
	}
	data := h.Result.(string)
	if len(data) < 578 {
		return "", errors.New("get auction fail, not found auction from chain")
	}
	if seller := util.AddHex(data[26:66]); seller == ethNoneAddress {
		return "", errors.New("get auction fail, not found seller form chain")
	}
	return data, err

}

func (c ethCall) BalanceOfBatch(contractAddr string, owners []string, ids []*big.Int) []uint64 {
	b, err := os.ReadFile(abiPath() + "contract/erc1155.abi")
	util.Panic(err)
	contract, err := c.EthRPC().Eth.NewContract(string(b))
	if err != nil {
		return nil
	}
	transParam := dto.TransactionParameters{To: contractAddr, Data: ""}
	if h, err := contract.Call(&transParam, "balanceOfBatch", owners, ids); err != nil || h.Error != nil {
		return nil
	} else {
		if raws := util.LogAnalysis(h.Result.(string)); len(raws) > 2 {
			var res []uint64
			for _, raw := range raws[2:] {
				res = append(res, util.U256(raw).Uint64())
			}
			return res
		}
	}
	return nil
}

func (c ethCall) SendRawTransaction(signedDat string) (string, error, int) {
	pointer := &dto.RequestResult{}
	params := make([]string, 1)
	params[0] = signedDat
	err := c.EthRPC().Provider.SendRequest(&pointer, "eth_sendRawTransaction", params)
	if err != nil {
		return "", err, -9999
	}
	if pointer.Error != nil {
		return "", errors.New(pointer.Error.Message), pointer.Error.Code
	}
	tx, _ := pointer.ToString()
	return tx, nil, 0
}

func (c ethCall) TakeBackNFTNonce(address string) (int64, error) {
	takeBackContract := util.GetContractAddress("takeBackNFT", c.Network)
	b, err := os.ReadFile("contract/takeBackNFT.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: takeBackContract, Data: ""}
	if h, err := contract.Call(&transParam, "userToNonce", address); err != nil {
		return 0, err
	} else {
		resultStr := h.Result.(string)
		return types.ComplexIntResponse(resultStr).ToInt64(), nil
	}
}

func (c ethCall) KeyStoreNonce(address string) (int64, error) {
	takeBackContract := util.GetContractAddress("rolesUpdater", c.Network)
	b, err := os.ReadFile("contract/rolesUpdater.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: takeBackContract, Data: ""}
	if h, err := contract.Call(&transParam, "userToNonce", address); err != nil {
		return 0, err
	} else {
		resultStr := h.Result.(string)
		return types.ComplexIntResponse(resultStr).ToInt64(), nil
	}
}

// StakingRewards.sol
func (c ethCall) RewardsToken(pool string) (string, error) {
	b, err := os.ReadFile("contract/StakingRewards.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: pool, Data: ""}
	if h, err := contract.Call(&transParam, "rewardsToken"); err != nil {
		return "", err
	} else {
		token := util.TrimHex(h.Result.(string))
		if len(token) < 64 {
			return "", errors.New("result length must be 64 bytes")
		}
		return util.AddHex(token[24:64]), nil
	}
}

func (c ethCall) StakingToken(pool string) (string, error) {
	b, err := os.ReadFile("contract/StakingRewards.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: pool, Data: ""}
	if h, err := contract.Call(&transParam, "stakingToken"); err != nil {
		return "", err
	} else {
		return util.AddHex(util.TrimHex(h.Result.(string))[24:64]), nil
	}
}

func (c ethCall) PeriodFinish(pool string) (int64, error) {
	b, err := os.ReadFile("contract/StakingRewards.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: pool, Data: ""}
	if h, err := contract.Call(&transParam, "periodFinish"); err != nil {
		return 0, err
	} else {
		b, err := h.ToBigInt()
		if err != nil {
			return 0, err
		}
		return b.Int64(), nil
	}
}

func (c ethCall) RewardRate(pool string) (*big.Int, error) {
	b, err := os.ReadFile("contract/StakingRewards.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: pool, Data: ""}
	if h, err := contract.Call(&transParam, "rewardRate"); err != nil {
		return big.NewInt(0), err
	} else {
		return h.ToBigInt()
	}
}

// UniswapV2Pair.sol
func (c ethCall) PairBalanceOf(lpToken string, pool string) (*big.Int, error) {
	b, err := os.ReadFile("contract/UniswapV2Pair.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: lpToken, Data: ""}
	if h, err := contract.Call(&transParam, "balanceOf", pool); err != nil {
		return big.NewInt(0), err
	} else {
		return h.ToBigInt()
	}
}

func (c ethCall) TotalSupply(lpToken string) (*big.Int, error) {
	b, err := os.ReadFile("contract/UniswapV2Pair.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: lpToken, Data: ""}
	if h, err := contract.Call(&transParam, "totalSupply"); err != nil {
		return big.NewInt(0), err
	} else {
		return h.ToBigInt()
	}
}

func (c ethCall) Token0(lpToken string) (string, error) {
	b, err := os.ReadFile("contract/UniswapV2Pair.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: lpToken, Data: ""}
	if h, err := contract.Call(&transParam, "token0"); err != nil {
		return "", err
	} else {
		return util.AddHex(util.TrimHex(h.Result.(string))[24:64]), nil
	}

}

func (c ethCall) Token1(lpToken string) (string, error) {
	b, err := os.ReadFile("contract/UniswapV2Pair.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: lpToken, Data: ""}
	if h, err := contract.Call(&transParam, "token1"); err != nil {
		return "", err
	} else {
		return util.AddHex(util.TrimHex(h.Result.(string))[24:64]), nil
	}
}

func (c ethCall) GetReserves(lpToken string) (reserve0 *big.Int, reserve1 *big.Int, blockTimestampLast int64, err error) {
	b, err := os.ReadFile("contract/UniswapV2Pair.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: lpToken, Data: ""}
	if h, err := contract.Call(&transParam, "getReserves"); err != nil {
		return big.NewInt(0), big.NewInt(0), 0, err
	} else {
		slice := util.LogAnalysis(h.Result.(string))
		reserve0 = util.EncodeU256(slice[0])
		reserve1 = util.EncodeU256(slice[1])
		blockTimestampLast = util.EncodeU256(slice[2]).Int64()
		return reserve0, reserve1, blockTimestampLast, nil
	}
}

func (c ethCall) ProtectPeriod(address, tokenId string) int64 {
	landContract := util.GetContractAddress("landResource", c.Network)
	b, err := os.ReadFile(abiPath() + "contract/landResource.abi")
	util.Panic(err)
	contract, err := c.EthRPC().Eth.NewContract(string(b))
	if err != nil {
		return 0
	}
	transParam := dto.TransactionParameters{To: landContract, Data: ""}
	if h, err := contract.Call(&transParam, "protectPeriod", address, util.U256(tokenId)); err != nil || h.Error != nil {
		return 0
	} else {
		b, err := h.ToBigInt()
		if err != nil {
			return 0
		}
		return b.Int64()
	}
}

func (c ethCall) TokenId2Apostle(tokenId string) []string {
	landContract := util.GetContractAddress("apostle", c.Network)
	b, err := os.ReadFile(abiPath() + "contract/apostle.abi")
	util.Panic(err)
	contract, _ := c.EthRPC().Eth.NewContract(string(b))
	transParam := dto.TransactionParameters{To: landContract, Data: ""}
	if h, err := contract.Call(&transParam, "tokenId2Apostle", util.U256(tokenId)); err != nil || h.Error != nil {
		return nil
	} else {
		data := h.Result.(string)
		return util.LogAnalysis(data)
	}
}
