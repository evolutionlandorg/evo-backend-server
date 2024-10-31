package storage

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/evolutionlandorg/block-scan/services"
	evoServices "github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/shopspring/decimal"
)

type tronCall struct {
	Call
	rpc *evoServices.TronClient
}

func (t tronCall) GetTokenLocationHM(tokenId string) (int64, int64, error) {
	res := t.rpc.ContractCall(util.GetContractAddress("tokenLocation", Tron), "",
		0, 0, "getTokenLocationHM(uint256)", tokenId)

	if res == nil || len(res.ConstantResult) == 0 {
		return 0, 0, errors.New("GetTokenLocationHM error")
	}
	slice := util.LogAnalysis(res.ConstantResult[0])
	x := util.EncodeU256(slice[0]).Int64()
	y := util.EncodeU256(slice[1]).Int64()
	return x, y, nil
}

func (t tronCall) OwnerOf(tokenId string) (string, error) {
	res := t.rpc.ContractCall(util.GetContractAddress("objectOwnership", Tron), "", 0, 0, "ownerOf(uint256)", tokenId)
	if res == nil || len(res.ConstantResult) == 0 || res.ConstantResult[0] == "" {
		return "", errors.New("GetTokenOwnerOf error")
	}
	owner := res.ConstantResult[0]
	return util.AddHex(util.TrimHex(owner)[24:64], Tron), nil
}

func (t tronCall) GetLandData(tokenId string) (string, error) {
	res := t.rpc.ContractCall(util.GetContractAddress("land", Tron), "", 0, 0, "getResourceRateAttr(uint256)", tokenId)
	if res == nil || len(res.ConstantResult) == 0 || res.ConstantResult[0] == "" {
		return "", errors.New("GetLandData error")
	}
	return res.ConstantResult[0], nil
}

func (t tronCall) ReceiptLog(tx string) (*services.Receipts, error) {
	tronReceipt := t.rpc.GetTxReceiptByHash(tx)
	if tronReceipt == nil {
		return nil, errors.New("nil receipt")
	}
	return evoServices.BuildReceipt(tronReceipt), nil
}

func (t tronCall) BlockNumber() uint64 {
	return t.rpc.BlockNumber()
}

func (t tronCall) FilterTrans(blockNum uint64, filter []string) (txn []string, contracts []string, timestamp uint64, transactionTo []string) {
	var block = t.rpc.GetBlockByNumber(blockNum)
	if block == nil {
		return
	}

	for _, tx := range block.Transactions {
		for _, c := range tx.RawData.Contract {
			if c.Type == "TriggerSmartContract" && util.StringInSlice(c.Parameter.Value.ContractAddress, filter) {
				contracts = append(contracts, c.Parameter.Value.ContractAddress)
				txn = append(txn, tx.TxID)
				break
			}
		}
	}
	timestamp = block.BlockHeader.RawData.Timestamp / 1000
	return
}

func (t tronCall) GetBalance(address string) *big.Int {
	return t.rpc.Balance(address)
}

func (t tronCall) BalanceOf(address, contract string) *big.Int {
	if res := t.rpc.ContractCall(contract, address, 0, 0, "balanceOf(address)", util.TrimTronHex(address)); res != nil && len(res.ConstantResult) > 0 {
		return util.U256(res.ConstantResult[0])
	}
	return big.NewInt(0)
}

func (t tronCall) UserToNonce(address, contractAddress string) int64 {
	res := t.rpc.ContractCall(contractAddress, "", 0, 0, "userToNonce(address)", util.TrimTronHex(address))
	if res == nil || len(res.ConstantResult) == 0 {
		return 0
	}
	return util.U256(res.ConstantResult[0]).Int64()
}

func (t tronCall) QuerySwapFee(amount *big.Int) decimal.Decimal {
	res := t.rpc.ContractCall(util.GetContractAddress("tokenSwap", Tron), "", 0, 0, "querySwapFee(uint256)", util.IntToHex(amount))
	if res == nil || len(res.ConstantResult) == 0 {
		return decimal.Zero
	}
	slice := util.LogAnalysis(res.ConstantResult[0])
	if len(slice) < 1 {
		return decimal.New(100, 0).Mul(decimal.NewFromFloat(1e18))
	}
	return decimal.NewFromBigInt(util.U256(slice[0]), 0)
}

func (t tronCall) PointsBalanceOf(address string) decimal.Decimal {
	res := t.rpc.ContractCall(util.GetContractAddress("userPoints", Tron), "", 0, 0, "pointsBalanceOf(address)", address)
	if res == nil || len(res.ConstantResult) == 0 || res.ConstantResult[0] == "" {
		return decimal.Zero
	}
	return util.BigToDecimal(util.U256(res.ConstantResult[0]))
}

func (t tronCall) TotalRewardInPool() decimal.Decimal {
	res := t.rpc.ContractCall(util.GetContractAddress("pointsReward", Tron),
		"", 0, 0, "totalRewardInPool(address)", util.GetContractAddress("ring", Tron))
	if res == nil || len(res.ConstantResult) == 0 || res.ConstantResult[0] == "" {
		return decimal.Zero
	}
	price := util.BigToDecimal(util.U256(res.ConstantResult[0]))
	return price.Round(5)
}

func (t tronCall) ComputePenalty(depositId int64) string {
	contractAddress := util.GetContractAddress("bank", Tron)
	res := t.rpc.ContractCall(contractAddress, "", 0, 0, "computePenalty(uint256)", util.IntToHex(depositId))
	if res == nil || len(res.ConstantResult) == 0 {
		return ""
	}
	return util.U256(res.ConstantResult[0]).String()
}

func (t tronCall) BlockHeader(blockNum uint64) *services.BlockHeader {
	var block = t.rpc.GetBlockByNumber(blockNum)
	if block == nil {
		return nil
	}
	header := services.BlockHeader{BlockTimeStamp: block.BlockHeader.RawData.Timestamp}
	return &header
}

func (t tronCall) GetTransactionStatus(tx string) string {
	trans := t.rpc.GetTransactionByHash(tx)
	if trans == nil {
		return ""
	}
	if len(trans.Ret) <= 0 {
		return ""
	}
	if trans.Ret[0].ContractRet == "SUCCESS" {
		return "Success"
	}
	return "Fail"
}

func (t tronCall) GetTransaction(tx string) *Transaction {
	res, _ := t.ReceiptLog(tx)
	if res == nil {
		return nil
	}
	txn := Transaction{BlockNum: util.U256(res.BlockNumber).Uint64()}
	return &txn
}

func (t tronCall) ApostleCurrentPriceInToken(tokenId string) decimal.Decimal {
	res := t.rpc.ContractCall(util.GetContractAddress("clockAuctionApostle", Tron), "", 0, 0, "getCurrentPriceInToken(uint256)", tokenId)
	if res == nil || len(res.ConstantResult) == 0 {
		return decimal.RequireFromString("-1")
	}
	return util.BigToDecimal(util.U256(res.ConstantResult[0]))
}

func (t tronCall) LandCurrentPriceInToken(tokenId string) decimal.Decimal {
	res := t.rpc.ContractCall(util.GetContractAddress("clockAuction", Tron), "", 0, 0, "getCurrentPriceInToken(uint256)", tokenId)
	if res == nil || len(res.ConstantResult) == 0 {
		return decimal.RequireFromString("-1")
	}
	return util.BigToDecimal(util.U256(res.ConstantResult[0]))
}

func (t tronCall) LandMask(tokenId string) (int, error) {
	res := t.rpc.ContractCall(util.GetContractAddress("land", Tron), "", 0, 0, "getFlagMask(uint256)", tokenId)
	if res == nil || len(res.ConstantResult) == 0 || res.ConstantResult[0] == "" {
		return 0, errors.New("GetLandMask error")
	}
	return util.StringToInt(util.U256(res.ConstantResult[0]).String()), nil

}

func (t tronCall) UnclaimedResource(tokenId string, resourceAddress []string) []string {
	resourceAddrStr := ""
	for _, addr := range resourceAddress {
		resourceAddrStr += fmt.Sprintf("%064s", addr[2:])
	}
	res := t.rpc.ContractCall(util.GetContractAddress("landResource", Tron), "", 0, 0, "availableResources(uint256,address[5])", tokenId, resourceAddrStr)
	if res == nil || len(res.ConstantResult) == 0 || res.ConstantResult[0] == "" {
		return nil
	}
	return util.LogAnalysis(res.ConstantResult[0])
}

func (c tronCall) DrillUnclaimedResource(drillTokenId string, resourceAddress []string) []string {
	return nil
}

func (t tronCall) Auction(tokenId string) (string, error) {
	res := t.rpc.ContractCall(util.GetContractAddress("clockAuction", Tron), "", 0, 0, "getAuction(uint256)", tokenId)
	if res == nil || len(res.ConstantResult) == 0 {
		return "", errors.New("getAuction error")
	}
	return res.ConstantResult[0], nil
}

// StakingRewards.sol
func (t tronCall) RewardsToken(pool string) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (t tronCall) StakingToken(pool string) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (t tronCall) PeriodFinish(pool string) (int64, error) {
	panic("not implemented") // TODO: Implement
}

func (t tronCall) RewardRate(pool string) (*big.Int, error) {
	panic("not implemented") // TODO: Implement
}

func (t tronCall) PairBalanceOf(lpToken, pool string) (*big.Int, error) {
	panic("not implemented") // TODO: Implement
}

func (t tronCall) TotalSupply(lpToken string) (*big.Int, error) {
	panic("not implemented") // TODO: Implement
}

func (t tronCall) Token0(lpToken string) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (t tronCall) Token1(lpToken string) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (t tronCall) GetReserves(lpToken string) (reserve0 *big.Int, reserve1 *big.Int, blockTimestampLast int64, err error) {
	panic("not implemented") // TODO: Implement
}
