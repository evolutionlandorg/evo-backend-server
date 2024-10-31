package models

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type AuctionHistory struct {
	gorm.Model
	AuctionId uint            `json:"auction_id"`
	TokenId   string          `json:"token_id"`
	TxId      string          `json:"tx_id"`
	Buyer     string          `json:"seller"`
	BidPrice  decimal.Decimal `json:"bid_price"  sql:"type:decimal(32,18);"`
	StartAt   int             `json:"start_at"`
	BidToken  string          `json:"bid_token"`
	AssetType string          `json:"asset_type"`
}

type AuctionHistoryJson struct {
	BidPrice decimal.Decimal `json:"bid_price"`
	TxId     string          `json:"tx_id"` // auction tx id
	Buyer    string          `json:"buyer"` // buyer address
	StartAt  int             `json:"start_at"`
	Name     string          `json:"name"`
}

type BidHistoryJson struct {
	TokenId      string          `json:"token_id"`
	BidPrice     decimal.Decimal `json:"bid_price"`
	TxId         string          `json:"tx_id"`
	Buyer        string          `json:"seller"`
	StartAt      int             `json:"start_at"`
	LastPrice    decimal.Decimal `json:"last_price" sql:"type:decimal(32,16);"`
	LastBidder   string          `json:"last_bidder"`
	LastBidStart int             `json:"last_bid_start"`
}

func (ah *AuctionHistory) New(ctx context.Context) error {
	if ah.exist(ctx) {
		return errors.New("tx exist")
	}
	result := util.WithContextDb(ctx).Create(&ah)
	return result.Error
}

func (ah *AuctionHistory) exist(ctx context.Context) bool {
	db := util.WithContextDb(ctx)
	one := AuctionHistory{}
	query := db.Where("tx_id = ? ", ah.TxId).First(&one)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return false
	}
	return true
}

func auctionHistoryList(ctx context.Context, auctionId uint, assetType, tokenId string) []AuctionHistoryJson {
	chain := GetChainByTokenId(tokenId)
	addressField := "wallet"
	switch chain {
	case TronChain:
		addressField = "tron_wallet"
	}
	db := util.WithContextDb(ctx)
	var ah []AuctionHistoryJson
	query := db.Table("auction_histories").
		Select("bid_price,tx_id,buyer,start_at,name").
		Joins("left join members on auction_histories.buyer = members."+addressField).
		Where("auction_id = ?", auctionId).
		Where("asset_type = ?", assetType).
		Order("auction_histories.bid_price desc").
		Scan(&ah)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	return ah
}

// 我的出价记录
func myBid(ctx context.Context, wallet, joinTables string) []BidHistoryJson {
	db := util.WithContextDb(ctx)
	var ah []BidHistoryJson
	query := db.Select("auction_histories.token_id,last_bid_start,last_bidder").
		Table("auction_histories").
		Where("buyer = ?", wallet).
		Where("status = ?", "going").
		Joins(fmt.Sprintf("join %s on auction_histories.auction_id = %s.id ", joinTables, joinTables)).
		Scan(&ah)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return ah
	}
	return ah
}

// 拍卖中的地块，并且用户有过出价记录，要么用户出价最高，等半小时后此地块进入未领取地块，要么出价被别人超过，用户可以点击继续出价进入拍卖页面
func BidingList(ctx context.Context, district int, wallet, op string, showPriceMap ...bool) ([]string, []string, map[string]decimal.Decimal) {
	var (
		claimTime                int64
		joinTables               string
		tokenIdArr, myLastBidArr []string
	)

	if op == AssetLand {
		claimTime = landClaimTime(district)
		joinTables = "auctions"
	} else if op == AssetApostle {
		claimTime = apostleClaimTime(district)
		joinTables = "auction_apostles"
	} else {
		return []string{}, []string{}, map[string]decimal.Decimal{}
	}

	history := myBid(ctx, wallet, joinTables)
	var (
		tokenIds = hashset.New()
	)
	for _, auction := range history {
		if auction.LastBidStart != 0 && time.Now().Unix()-int64(auction.LastBidStart) < claimTime {
			tokenIdArr = append(tokenIdArr, auction.TokenId)
			if strings.EqualFold(auction.LastBidder, strings.ToUpper(wallet)) {
				myLastBidArr = append(myLastBidArr, auction.TokenId)
				tokenIds.Add(auction.TokenId)
			}
		}
	}
	priceMap := make(map[string]decimal.Decimal)
	if tokenIds.Size() != 0 && len(showPriceMap) != 0 && showPriceMap[0] {
		auctions := FindAuctionsByTokenIds(ctx, tokenIds.Values())
		for _, auction := range auctions {
			priceMap[auction.TokenId] = auction.CurrentPriceLocal()
		}
	}
	return tokenIdArr, myLastBidArr, priceMap
}
