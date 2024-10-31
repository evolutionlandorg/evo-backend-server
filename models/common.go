package models

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/evolutionlandorg/block-scan/services"
	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/shopspring/decimal"
)

const (
	noneAddress     = "0x0000000000000000000000000000000000000000"
	tronNoneAddress = "410000000000000000000000000000000000000000"
	// noneTokenId       = "0000000000000000000000000000000000000000000000000000000000000000"
	// regReward         = "500"
	//firstLayerReward  = "150"
	//secondLayerReward = "50"

	goldBoxMin      = 15000
	silverBoxMin    = 5000
	silverBoxRandom = 1000
	goldBoxRandom   = 3000
	// GoldBoxPrice    = 1
	// SilverBoxPrice  = 1

	accountAdd = "add"
	accountSub = "sub"

	currencyRing = "ring"
	currencyCoo  = "coo"
	currencyEth  = "eth"
	// currencyTrx   = "trx"
	currencyKton  = "kton"
	currencyGold  = "gold"
	currencyWood  = "wood"
	currencyWater = "water"
	currencyFire  = "fire"
	currencySoil  = "soil"

	AssetLand        = "land"
	AssetApostle     = "apostle"
	AssetMirrorKitty = "mirrorKitty"
	AssetDrill       = "drill"
	AssetItem        = "item"
	AssetEquipment   = "equipment"
	Material         = "material"

	// reasonReg                  = "reg"
	reasonInviteFirst    = "inviteFirst"
	reasonSecondFirst    = "inviteSecond"
	reasonUnlockTreasure = "unlockTreasure"
	// reasonKtonDividend         = "ktonDividend"
	reasonSwap = "swap"
	// ReasonAirDrop              = "airDrop"
	ReasonRedPacket         = "redPacket"
	ReasonRedPacketSendBack = "redPacketSendBack"
	// ReasonApostleArenaSendBack = "apostleArenaSendBack"
	// ReasonApostleArenaWin      = "apostleArenaWin"
	// ReasonApostleArenaBetWin   = "apostleArenaBetWin"
	ReasonApostleAFK       = "ApostleAFK"
	ReasonBuyTicket        = "buyTicket"
	ReasonDungeonClearance = "dungeonClearance"
	ReasonWithdrawMaterial = "withdrawMaterial"

	landClaimReward        = "0"
	landClaimWaiting int64 = 1800
	landOnsell             = "onsell" // 拍卖状态
	landFresh              = "fresh"

	AuctionGoing   = "going"     // 拍卖中
	AuctionClaimed = "unclaimed" // 拍卖成功未领取的状态
	AuctionCancel  = "cancel"    // 取消拍卖
	AuctionFinish  = "finish"    // 拍卖完成
	AuctionOver    = "over"      // 打工结束

	TransactionHistoryWithdraw        = "withdraw"        // 提现
	TransactionHistoryLandSale        = "landSale"        // 拍卖
	TransactionHistoryCancelAuction   = "cancelAuction"   // 取消拍卖
	TransactionHistorySuccessAuction  = "successAuction"  // 领取地块
	TransactionHistoryBid             = "bid"             // 拍地
	TransactionHistoryBidRefund       = "bidRefund"       // 拍地退款
	TransactionHistoryUniswapEx       = "uniswapEx"       // uniswap
	TransactionHistoryDepositBank     = "depositBank"     // 存款
	TransactionHistoryDepositReward   = "depositReward"   // 存款奖励
	TransactionHistoryWithdrawBank    = "withdrawBank"    // 取款
	TransactionHistoryWithdrawPenalty = "withdrawPenalty" // 取款罚金
	TransactionHistoryTickets         = "tickets"         // 点数抽奖
	TransactionHistoryKtonMapping     = "ktonMapping"     // kton迁移
	TransactionHistoryRingMapping     = "ringMapping"     // ring迁移
	TransactionHistoryLuckyBox        = "luckyBox"        // luckyBox v2
	TransactionHistoryLuckyBoxOpen    = "luckyBoxOpen"    // luckyBox open
	TransactionHistoryNewbieReward    = "newbieReward"    // luckyBox open
	DrillCreate                       = "drillCreate"
	DrillEnchanced                    = "drillEnchanced"
	DrillDisenchanted                 = "drillDisenchanted"

	TransactionHistoryApostleSale             = "apostleSale"             // 使徒拍卖
	TransactionHistoryApostleCancelSale       = "apostleCancelSale"       // 使徒拍卖取消
	TransactionHistoryApostleBid              = "apostleBid"              // 拍使徒
	TransactionHistoryApostleSuccessAuction   = "apostleSuccessAuction"   // 领取使徒
	TransactionHistoryApostleBidRefund        = "apostleBidRefund"        // 拍使徒退款
	TransactionHistoryClaimResource           = "claim_resource"          // 提取地块资源
	TransactionHistorySwap                    = "swap"                    // swap
	TransactionHistoryApostleFertility        = "apostleFertility"        // 使徒繁殖
	TransactionHistoryApostleFertilityCancel  = "apostleFertilityCancel"  // 使徒繁殖
	TransactionHistoryApostleFertilitySuccess = "apostleFertilitySuccess" // 使徒繁殖
	TransactionHistoryApostlePregnant         = "apostlePregnant"         // 使徒生育
	TransactionHistoryApostleRent             = "apostleRent"             // 使徒出租
	TransactionHistoryApostleRentCancel       = "apostleRentCancel"       // 使徒出租取消
	TransactionHistoryApostleWorkingStop      = "apostleWorkingStop"      // 使徒打工取消
	TransactionHistoryApostleRentSuccess      = "apostleRentSuccess"      // 使徒出租成功
	TransactionHistoryApostleBindPet          = "apostleBindPet"
	TransactionHistoryApostleUnbindPet        = "apostleUnbindPet"
	TransactionHistoryPending                 = "pending"

	apostleClaimWaiting    = 600
	apostleFresh           = "fresh"
	apostleOnsell          = "onsell"
	apostleFertility       = "fertility"
	apostleBirth           = "birth"
	apostleRent            = "rent"
	apostleHiring          = "hiring"
	apostleWorking         = "working"
	apostlePve             = "pve"
	apostleBirthUnclaimed  = "birthUnclaimed"
	apostleHiringUnclaimed = "hiringUnclaimed"
	apostleBuild           = "build"
	apostleBuildAdmin      = "admin"

	apostleTradeTex = "0.96"
	fertilityPrice  = "50"

	TronChain    = "Tron"
	EthChain     = "Eth"
	CrabChain    = "Crab"
	HecoChain    = "Heco"
	PolygonChain = "Polygon"
	// BscChain     = "Bsc"

	BroadcastPointReward = "pointReward"
	BroadcastLandAuction = "landAuction"

	tronLandAttenuationAt = 1546560000
	ethLandAttenuationAt  = 1545580800

	AdventureGoing = "going"
	AdventureStop  = "stop"
)

type StatSum struct {
	Total decimal.Decimal `json:"total"`
}

type EthTransactionCallback struct {
	Tx             string
	Receipt        *services.Receipts
	BlockTimestamp int64 `json:"block_timestamp"`
}

type ListOpt struct {
	Page       int
	Row        int
	OrderField string
	Order      string
	Display    string
	Filter     string
	WhereQuery []interface{}
	Chain      string
}

func getStringValueByFieldName(n interface{}, fieldName string) (string, bool) {
	s := reflect.ValueOf(n)
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	if s.Kind() != reflect.Struct {
		return "", false
	}
	f := s.FieldByName(fieldName)
	if !f.IsValid() {
		return "", false
	}
	switch f.Kind() {
	case reflect.String:
		return f.Interface().(string), true
	case reflect.Int:
		return strconv.FormatInt(f.Int(), 10), true
	case reflect.Uint:
		return strconv.FormatUint(f.Uint(), 10), true
	case reflect.Int64:
		return strconv.FormatInt(f.Int(), 10), true
	default:
		if reflect.TypeOf(f.Interface()).Name() == "Decimal" {
			return f.Interface().(decimal.Decimal).String(), true
		}
		return "", false
	}
}

func GetChainByTokenId(tokenId string) string {
	chainId := tokenId[2:4]
	i, err := strconv.ParseUint(chainId, 16, 0)
	if err != nil {
		return EthChain
	}
	for chain, ChainId := range util.Evo.ChainId {
		if ChainId == uint(i) {
			return chain
		}
	}
	return EthChain
}

func GetChainById(id int) string {
	for chain, ChainId := range util.Evo.ChainId {
		if ChainId == uint(id) {
			return chain
		}
	}
	return EthChain
}

func GetDistrictByChain(chain string) int {
	if district, ok := util.Evo.District[chain]; ok {
		return int(district)
	} else {
		return 1
	}
}

func GetChainByDistrict(district int) string {
	for chain, v := range util.Evo.District {
		if district == int(v) {
			return chain
		}
	}
	return "Eth"
}

func getNFTDistrict(tokenId string) int {
	return GetDistrictByChain(GetChainByTokenId(tokenId))
}

func getAssetTypeByTokenId(tokenId string) string {
	b := util.EncodeU256(tokenId)
	if new(big.Int).And(new(big.Int).Rsh(b, 128), big.NewInt(65535)).Cmp(big.NewInt(256)) == 0 {
		return AssetMirrorKitty
	} else if new(big.Int).And(new(big.Int).Rsh(b, 192), big.NewInt(255)).Cmp(big.NewInt(1)) == 0 {
		return AssetLand
	} else if new(big.Int).And(new(big.Int).Rsh(b, 192), big.NewInt(255)).Cmp(big.NewInt(2)) == 0 {
		return AssetApostle
	} else if new(big.Int).And(new(big.Int).Rsh(b, 192), big.NewInt(255)).Cmp(big.NewInt(4)) == 0 {
		return AssetDrill
	} else if new(big.Int).And(new(big.Int).Rsh(b, 192), big.NewInt(255)).Cmp(big.NewInt(5)) == 0 {
		return AssetItem
	} else if new(big.Int).And(new(big.Int).Rsh(b, 192), big.NewInt(255)).Cmp(big.NewInt(6)) == 0 {
		return AssetEquipment
	} else if new(big.Int).And(new(big.Int).Rsh(b, 192), big.NewInt(255)).Cmp(big.NewInt(11)) == 0 {
		return Material
	} else {
		return ""
	}
}

func GetAssetTypeById(id int) string {
	assetType := map[int]string{
		256: AssetMirrorKitty,
		1:   AssetLand,
		2:   AssetApostle,
		4:   AssetDrill,
		5:   AssetItem,
		6:   AssetEquipment,
		7:   Material,
	}
	return assetType[id]
}

const (
	_ = iota
	// landContractId
	// ApostleContractId
	// IteringNft
	BuildingContractId
)

func interstellarEncoding(nftContractId, district, tokenIndex int) string {
	index := fmt.Sprintf("%032s", fmt.Sprintf("%x", tokenIndex+1))
	return fmt.Sprintf("2a0%d00010%d000101000000000000000%d%s", district, nftContractId, district, index)
}

// func paginationList(query *gorm.DB, opt *ListOpt) *gorm.DB {
//	if opt != nil {
//		if opt.OrderField != "" && opt.Order != "" {
//			query = query.Order(fmt.Sprintf("%s %s", opt.OrderField, opt.Order))
//		}
//		if opt.Row > 0 {
//			query = query.Offset(opt.Page * opt.Row).Limit(opt.Row)
//		}
//	}
//	return query
// }

func refreshOpenSeaMetadata(tokenId string) {
	if util.IsProduction() {
		go util.HttpGet(fmt.Sprintf("https://api.opensea.io/api/v1/asset/0x14a4123da9ad21b2215dc0ab6984ec1e89842c6d/%s/?force_update=true", util.U256(tokenId).String()))
	}
}

// temp change staging Eth claim time 300
func landClaimTime(district int) int64 {
	if (district == 4) && !util.IsProduction() {
		return 300
	}
	return landClaimWaiting
}

// temp change staging Eth claim time 300
func apostleClaimTime(district int) int64 {
	if (district == 4) && !util.IsProduction() {
		return 300
	}
	return apostleClaimWaiting
}
