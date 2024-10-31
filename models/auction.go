package models

import (
	"context"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type Auction struct {
	gorm.Model
	TokenId      string          `json:"token_id"`
	CreateTX     string          `json:"create_tx"`
	Seller       string          `json:"seller"`
	StartPrice   decimal.Decimal `json:"start_price" sql:"type:decimal(32,16);"`
	EndPrice     decimal.Decimal `json:"end_price" sql:"type:decimal(32,16);"`
	Duration     int             `json:"duration"`
	StartAt      int             `json:"start_at"`
	Status       string          `json:"status"` // going/unclaimed/finish/cancel
	Winner       string          `json:"winner"`
	FinalPrice   decimal.Decimal `json:"final_price" sql:"type:decimal(32,16);"`
	LastPrice    decimal.Decimal `json:"last_price" sql:"type:decimal(32,16);"`
	LastBidder   string          `json:"last_bidder"`
	LastBidStart int             `json:"last_bid_start"`
	Currency     string          `json:"currency"`
	ClaimTime    int             `json:"claim_time"`
	District     int             `json:"district"`
}

type AuctionJson struct {
	Seller          string               `json:"seller"`                               // seller address
	SellerName      string               `json:"seller_name"`                          // seller name
	Status          string               `json:"status" enums:"[cancel,finish,going]"` // auction status
	CurrentPrice    decimal.Decimal      `json:"current_price"`                        // current price
	Duration        int                  `json:"duration"`
	StartAt         int                  `json:"start_at"`
	StartPrice      decimal.Decimal      `json:"start_price" sql:"type:decimal(32,16);"` // start price
	EndPrice        decimal.Decimal      `json:"end_price" sql:"type:decimal(32,16);"`   // end price
	LastPrice       decimal.Decimal      `json:"last_price" sql:"type:decimal(32,16);"`
	LastBidStart    int                  `json:"last_bid_start"`
	History         []AuctionHistoryJson `json:"history"` // auction history
	CurrentTime     int64                `json:"current_time"`
	LandClaimReward decimal.Decimal      `json:"land_claim_reward"`
	ClaimWaiting    int                  `json:"claim_waiting"`
	WinnerName      string               `json:"winner_name"`    // winner name
	WinnerAddress   string               `json:"winner_address"` // winner address
	Token           *util.Token          `json:"token"`
}

type LandAuctionHistoryJson struct {
	Seller     string          `json:"seller"`
	FinalPrice decimal.Decimal `json:"final_price" sql:"type:decimal(32,16);"`
	Winner     string          `json:"winner"`
	CreateTX   string          `json:"create_tx"`
	ClaimTime  int             `json:"claim_time"`
}

// FindAuctionsByTokenIds 根据 tokenIds 获取 Auction
func FindAuctionsByTokenIds(ctx context.Context, tokenIds []interface{}, selectValue ...string) (auctions []Auction) {
	if len(tokenIds) == 0 {
		return
	}
	db := util.WithContextDb(ctx)
	db = db.Where("token_id IN (?)", tokenIds)
	if len(selectValue) != 0 {
		db = db.Select(strings.Join(selectValue, ","))
	}
	db.Find(&auctions)
	return
}

// CurrentPrice 从链上获取准确的价格, 如果不需要精准的价格可以使用 CurrentPriceLocal 函数从本地计算
func (auc *Auction) CurrentPrice() decimal.Decimal {
	chain := GetChainByTokenId(auc.TokenId)
	now := int(time.Now().Unix())

	if auc.Status != AuctionGoing || now < auc.StartAt || auc.Duration == 0 {
		return decimal.RequireFromString("-1")
	}
	sg := storage.New(chain)
	return sg.LandCurrentPriceInToken(auc.TokenId)
}

func GetCurrentAuction(ctx context.Context, tokenId, status string) *Auction {
	db := util.WithContextDb(ctx)
	var auc Auction
	var query *gorm.DB
	if status != "" {
		query = db.Where("token_id = ? and status = ?", tokenId, status).First(&auc)
	} else {
		query = db.Where("token_id = ?", tokenId).First(&auc)
	}
	if query.RecordNotFound() {
		return nil
	}
	return &auc
}

func MyAuctionLandList(ctx context.Context, district int, wallets []string) (tokenIds []string) {
	db := util.WithContextDb(ctx)
	var query *gorm.DB

	remaining := time.Now().Unix() - landClaimTime(district)
	if len(wallets) < 1 {
		none := noneAddress
		if district == 2 {
			none = tronNoneAddress
		}
		query = db.Table("auctions").Where("district = ? and status = ? and last_bidder != ? and last_bid_start < ? ", district, AuctionGoing, none, remaining).Pluck("token_id", &tokenIds)
	} else {
		query = db.Table("auctions").Where("district = ? and status = ? and last_bidder = ? and last_bid_start <?", district, AuctionGoing, wallets[0], remaining).Pluck("token_id", &tokenIds)
	}

	if query.RecordNotFound() {
		return
	}
	return tokenIds
}

func OnsellLandList(ctx context.Context, district int, where string) (tokenIdArr []string, hasBidArr []string, priceMap map[string]decimal.Decimal, startAtMap map[string]int, tokenMap map[string]*util.Token) {
	db := util.WithContextDb(ctx)
	var aucs []Auction
	priceMap = make(map[string]decimal.Decimal)
	tokenMap = make(map[string]*util.Token)
	startAtMap = make(map[string]int)
	query := db.Where("district = ?", district).Where(where).Find(&aucs)
	if query.RecordNotFound() {
		return
	}
	for _, auction := range aucs {
		tokenMap[auction.TokenId] = util.Evo.GetToken(GetChainByDistrict(auction.District), auction.Currency)
		priceMap[auction.TokenId] = auction.CurrentPriceLocal()
		startAtMap[auction.TokenId] = auction.StartAt
		if auction.LastBidStart != 0 && time.Now().Unix()-int64(auction.LastBidStart) > landClaimTime(district) { // 排除unclaimed
			continue
		}
		tokenIdArr = append(tokenIdArr, auction.TokenId)
		if auction.LastBidStart != 0 {
			hasBidArr = append(hasBidArr, auction.TokenId)
		}
	}
	return
}

func landAuctionHistory(ctx context.Context, tokenId string) []LandAuctionHistoryJson {
	db := util.WithContextDb(ctx)
	var lah []LandAuctionHistoryJson
	query := db.Table("auctions").Where("token_id = ?", tokenId).Where("status = ?", AuctionFinish).Order("id desc").Scan(&lah)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	return lah
}

func notStartLands(ctx context.Context) []string {
	db := util.WithContextDb(ctx)
	var tokenIdArr []string
	query := db.Table("auctions").Where("status = ? and start_at > ?", AuctionGoing, int(time.Now().Unix())).Pluck("token_id", &tokenIdArr)
	if query.RecordNotFound() {
		return tokenIdArr
	}
	return tokenIdArr
}

func genesisLands(ctx context.Context, chain string) []string {
	db := util.WithContextDb(ctx)
	var tokenIdArr []string
	genesisHolder := util.GetContractAddress("genesisHolder", chain)
	query := db.Table("auctions").Where("district = ? and status = ? and start_at < ? and seller = ?", GetDistrictByChain(chain), AuctionGoing, int(time.Now().Unix()), genesisHolder).Pluck("token_id", &tokenIdArr)
	if query.RecordNotFound() {
		return tokenIdArr
	}
	return tokenIdArr
}

func (auc *Auction) CurrentPriceLocal() decimal.Decimal {
	now := int(time.Now().Unix())
	if auc == nil || auc.Status != AuctionGoing || now < auc.StartAt || auc.Duration == 0 {
		return decimal.RequireFromString("-1")
	}
	if auc.LastBidder == noneAddress || auc.LastBidder == "" || auc.LastBidder == tronNoneAddress {
		secondsPassed := now - auc.StartAt
		if secondsPassed > auc.Duration {
			return auc.EndPrice
		} else {
			totalPriceInRINGChange := auc.EndPrice.Sub(auc.StartPrice)
			secondsDecimal := decimal.RequireFromString(util.IntToString(secondsPassed))
			durationDecimal := decimal.RequireFromString(util.IntToString(auc.Duration))
			currentPriceInRINGChange := totalPriceInRINGChange.Mul(secondsDecimal.Div(durationDecimal))
			return auc.StartPrice.Add(currentPriceInRINGChange).Round(18)
		}
	} else {
		return auc.LastPrice.Mul(decimal.NewFromFloat(1.1)).Round(18)
	}
}

func (auc *Auction) AsJson(ctx context.Context) *AuctionJson {
	currencyRingPrice := auc.CurrentPrice()
	aj := AuctionJson{
		Seller:          auc.Seller,
		Status:          auc.Status,
		CurrentPrice:    currencyRingPrice,
		Duration:        auc.Duration,
		StartAt:         auc.StartAt,
		StartPrice:      auc.StartPrice,
		EndPrice:        auc.EndPrice,
		LastPrice:       auc.LastPrice,
		LastBidStart:    auc.LastBidStart,
		History:         auctionHistoryList(ctx, auc.ID, AssetLand, auc.TokenId),
		CurrentTime:     time.Now().Unix(),
		LandClaimReward: decimal.RequireFromString(landClaimReward),
		ClaimWaiting:    int(landClaimTime(auc.District)),
		Token:           util.Evo.GetToken(GetChainByDistrict(auc.District), auc.Currency),
	}

	member := GetMemberByAddress(ctx, auc.Seller, GetChainByTokenId(auc.TokenId))
	if member != nil {
		aj.SellerName = member.Name
	}
	if aj.LastBidStart != 0 && time.Now().Unix()-int64(aj.LastBidStart) > int64(aj.ClaimWaiting) {
		aj.Status = AuctionClaimed
		winner := GetMemberByAddress(ctx, auc.LastBidder, GetChainByTokenId(auc.TokenId))
		if winner != nil {
			aj.WinnerName = winner.Name
			aj.WinnerAddress = auc.LastBidder
		}
	}
	return &aj
}
