package storage

import (
	"math/big"

	"github.com/evolutionlandorg/block-scan/services"
	evoServices "github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/roundrobin"

	"github.com/evolutionlandorg/evo-backend/pkg/github.com/regcostajr/go-web3"

	"github.com/shopspring/decimal"
)

var (
	ethNoneAddress = "0x0000000000000000000000000000000000000000"
)

type IStorage interface {
	// RPC
	ReceiptLog(tx string) (*services.Receipts, error)
	BlockNumber() uint64
	FilterTrans(blockNum uint64, filter []string) (txn []string, contracts []string, timestamp uint64, transactionTo []string)
	GetBalance(address string) *big.Int
	BlockHeader(blockNum uint64) *services.BlockHeader
	GetTransactionStatus(tx string) string
	GetTransaction(tx string) *Transaction
	SendRawTransaction(signedDat string) (string, error, int)

	// Contract
	GetTokenLocationHM(tokenId string) (int64, int64, error)
	OwnerOf(tokenId string) (string, error)
	GetResourceRateAttr(tokenId string) (string, error)
	DEGOOwnerOf(tokenId string) string
	DEGOTokensOfOwner(owner string) []string
	UserToNonce(address, contract string) int64
	BalanceOf(address, contract string) *big.Int
	QuerySwapFee(amount *big.Int) decimal.Decimal
	TotalRewardInPool() decimal.Decimal
	PointsBalanceOf(address string) decimal.Decimal
	ComputePenalty(depositId int64) string
	ApostleCurrentPriceInToken(tokenId string) decimal.Decimal
	LandCurrentPriceInToken(tokenId string) decimal.Decimal
	DrillUnclaimedResource(drillTokenId string, resourceAddress []string) []string
	LandMask(tokenId string) (int, error)
	UnclaimedResource(tokenId string, resourceAddress []string) []string
	Auction(tokenId string) (string, error)
	ProtectPeriod(address, tokenId string) int64
	// ERC1155
	BalanceOfBatch(string, []string, []*big.Int) []uint64

	// StakingRewards.sol
	RewardsToken(pool string) (string, error)
	StakingToken(pool string) (string, error)
	PeriodFinish(pool string) (int64, error)
	RewardRate(pool string) (*big.Int, error)

	// UniswapV2Pair.sol
	PairBalanceOf(lpToken, pool string) (*big.Int, error)
	TotalSupply(lpToken string) (*big.Int, error)
	Token0(lpToken string) (string, error)
	Token1(lpToken string) (string, error)
	GetReserves(lpToken string) (reserve0, reserve1 *big.Int, blockTimestampLast int64, err error)

	// Apostle
	TokenId2Apostle(tokenId string) []string
}

type BlockHeader struct {
	BlockTimeStamp uint64
	Hash           string
}

type Transaction struct {
	BlockNum uint64
}

type Call struct {
	Network   string
	RpcUrl    string
	RpcClient *roundrobin.Balancer
}

// StakingRewards.sol
func (c Call) RewardsToken(pool string) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (c Call) StakingToken(pool string) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (c Call) PeriodFinish(pool string) (int64, error) {
	panic("not implemented") // TODO: Implement
}

func (c Call) RewardRate(pool string) (*big.Int, error) {
	panic("not implemented") // TODO: Implement
}

func (c Call) PairBalanceOf(lpToken, pool string) (*big.Int, error) {
	panic("not implemented") // TODO: Implement
}

func (c Call) TotalSupply(lpToken string) (*big.Int, error) {
	panic("not implemented") // TODO: Implement
}

func (c Call) Token0(lpToken string) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (c Call) Token1(lpToken string) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (c Call) GetReserves(lpToken string) (reserve0 *big.Int, reserve1 *big.Int, blockTimestampLast int64, err error) {
	panic("not implemented") // TODO: Implement
}

func (c Call) SendRawTransaction(signedDat string) (string, error, int) {
	panic("implement me")
}

func (c Call) BalanceOfBatch(string, []string, []*big.Int) []uint64 {
	panic("implement me")
}

func (c Call) LandMask(tokenId string) (int, error) {
	panic("implement me")
}

func (c Call) UnclaimedResource(tokenId string, resourceAddress []string) []string {
	panic("implement me")
}

func (c Call) Auction(tokenId string) (string, error) {
	panic("implement me")
}

func (c Call) LandCurrentPriceInToken(tokenId string) decimal.Decimal {
	panic("implement me")
}

func (c Call) ApostleCurrentPriceInToken(tokenId string) decimal.Decimal {
	panic("implement me")
}

func (c Call) GetTransaction(tx string) *Transaction {
	panic("implement me")
}

func (c Call) GetTransactionStatus(tx string) string {
	panic("implement me")
}

func (c Call) BlockHeader(uint64) *services.BlockHeader {
	panic("implement me")
}

func (c Call) ComputePenalty(depositId int64) string {
	panic("implement me")
}

func (c Call) PointsBalanceOf(address string) decimal.Decimal {
	panic("implement me")
}

func (c Call) TotalRewardInPool() decimal.Decimal {
	panic("implement me")
}

func (c Call) QuerySwapFee(amount *big.Int) decimal.Decimal {
	panic("implement me")
}

func (c Call) UserToNonce(address, contract string) int64 {
	panic("implement me")
}

func (c Call) BalanceOf(address, contract string) *big.Int {
	panic("implement me")
}

func (c Call) GetBalance(address string) *big.Int {
	panic("implement me")
}

func (c Call) DEGOOwnerOf(tokenId string) string {
	panic("implement me")
}

func (c Call) FilterTrans(blockNum uint64, filter []string) (txn []string, contracts []string, timestamp uint64, transactionTo []string) {
	panic("implement me")
}

func (c Call) BlockNumber() uint64 {
	return 0
}

func (c Call) ReceiptLog(tx string) (*services.Receipts, error) {
	return nil, nil
}

func (c Call) DEGOTokensOfOwner(owner string) []string {
	return nil
}

func (c Call) GetResourceRateAttr(tokenId string) (string, error) {
	return "", nil
}

func (c Call) OwnerOf(string) (string, error) {
	return "", nil
}

func (c Call) GetTokenLocationHM(string) (int64, int64, error) {
	return 0, 0, nil
}

func (c Call) ProtectPeriod(address, tokenId string) int64 {
	return 0
}

func (c Call) TokenId2Apostle(tokenId string) []string {
	return nil
}

func (c Call) DrillUnclaimedResource(drillTokenId string, resourceAddress []string) []string {
	return nil
}

func (c Call) EthRPC() *web3.Web3 {
	item, _ := c.RpcClient.Pick()
	return item.(*web3.Web3)
}

const (
	Ethereum = "Eth"
	Tron     = "Tron"
	Crab     = "Crab"
	Heco     = "Heco"
	Bsc      = "Bsc"
	Polygon  = "Polygon"
)

func GetChainWssRpc(chain string) string {
	switch chain {
	case Ethereum:
		return util.GetEnv("ETH_WSS_RPC", "")
	case Crab:
		return util.GetEnv("CRAB_WSS_RPC", "wss://darwinia-crab.api.onfinality.io/public-ws")
	case Tron:
		return ""
	case Heco:
		return util.GetEnv("HECO_WSS_RPC", "wss://ws-testnet.hecochain.com")
	case Polygon:
		return util.GetEnv("POLYGON_WSS_RPC", "wss://polygon-mumbai.api.onfinality.io/public")
	}
	return ""
}

func New(chain string) IStorage {
	c := Call{Network: chain}
	switch chain {
	case Ethereum:
		c.RpcClient = util.Rb
		return ethCall{Call: c}
	case Crab:
		crab := crabCall{ethCall{Call: c}}
		crab.initClient()
		return crab
	case Tron:
		rpc := evoServices.TronClient{TrxRPC: util.TronRpc}
		return tronCall{Call: c, rpc: &rpc}
	case Heco:
		heco := hecoCall{ethCall{Call: c}}
		heco.initClient()
		return heco
	case Bsc:
		bsc := bscCall{ethCall{Call: c}}
		bsc.initClient()
		return bsc
	case Polygon:
		Polygon := polygonCall{ethCall{Call: c}}
		Polygon.initClient()
		return Polygon
	}
	return c
}

func abiPath() string {
	return util.GetEnv("ABI_PATH", "")
}
