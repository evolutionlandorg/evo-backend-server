package models

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type AuctionApostle struct {
	gorm.Model
	ApostleId    uint            `json:"apostle_id"`
	TokenId      string          `json:"token_id"`
	CreateTX     string          `json:"create_tx"`
	Seller       string          `json:"seller"`
	StartPrice   decimal.Decimal `json:"start_price" sql:"type:decimal(36,18);"`
	EndPrice     decimal.Decimal `json:"end_price" sql:"type:decimal(36,18);"`
	Duration     int             `json:"duration"`
	StartAt      int             `json:"start_at"`
	Status       string          `json:"status"` // going/unclaimed/finish/cancel
	Winner       string          `json:"winner"`
	FinalPrice   decimal.Decimal `json:"final_price" sql:"type:decimal(36,18);"`
	LastPrice    decimal.Decimal `json:"last_price" sql:"type:decimal(36,18);"`
	LastBidder   string          `json:"last_bidder"`
	LastBidStart int             `json:"last_bid_start"`
	Currency     string          `json:"currency"`
	ClaimTime    int             `json:"claim_time"`
	District     int             `json:"district"`
}

type ApostleAuctionSample struct {
	StartAt      int                  `json:"start_at"`
	Duration     int                  `json:"duration"`
	Status       string               `json:"status"`
	Seller       string               `json:"seller"`
	SellerName   string               `json:"seller_name"`
	StartPrice   decimal.Decimal      `json:"start_price"`
	EndPrice     decimal.Decimal      `json:"end_price"`
	LastPrice    decimal.Decimal      `json:"last_price"`
	LastBidStart int                  `json:"last_bid_start"`
	ClaimWaiting int                  `json:"claim_waiting"`
	WinnerName   string               `json:"winner_name"`
	Winner       string               `json:"winner"`
	CurrentPrice decimal.Decimal      `json:"current_price"`
	CurrentTime  int64                `json:"current_time"`
	History      []AuctionHistoryJson `json:"history"`
	Token        *util.Token          `json:"token"`
}

func (auc *AuctionApostle) New(db *util.GormDB) error {
	auc.Status = AuctionGoing
	result := db.Create(&auc)
	return result.Error
}

func (auc *AuctionApostle) CurrentPrice() decimal.Decimal {
	now := int(time.Now().Unix())
	if auc == nil || auc.Status != AuctionGoing || now < auc.StartAt || auc.Duration == 0 {
		return decimal.RequireFromString("-1")
	}
	if auc.LastBidder == noneAddress || auc.LastBidder == "" {
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

func (auc *AuctionApostle) CurrentPriceFromChain() decimal.Decimal {
	chain := GetChainByTokenId(auc.TokenId)
	now := int(time.Now().Unix())
	if auc.Status != AuctionGoing || now < auc.StartAt || auc.Duration == 0 {
		return decimal.RequireFromString("-1")
	}
	if (auc.LastBidder == noneAddress || auc.LastBidder == "") && (now-auc.StartAt > auc.Duration) {
		return auc.EndPrice
	}
	sg := storage.New(chain)
	return sg.ApostleCurrentPriceInToken(auc.TokenId)
}

// ClockAuctionApostleCallback 使徒拍卖/取消/拍卖成功/出价回调函数
func (ec *EthTransactionCallback) ClockAuctionApostleCallback(ctx context.Context) (err error) {
	if getTransactionDeal(ctx, ec.Tx, "clockAuctionApostle") != nil {
		return errors.New("tx exist")
	}
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	var event, winner, tokenId string
	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("clockAuctionApostle", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			case services.AbiEncodingMethod("AuctionCreated(uint256,address,uint256,uint256,uint256,address,uint256)"):
				dataSlice := util.LogAnalysis(log.Data)
				err = createAuctionApostle(ctx, db, dataSlice, ec.Tx, chain)
			case services.AbiEncodingMethod("AuctionCancelled(uint256)"):
				err = cancelAuctionApostle(ctx, db, ec.Tx, util.TrimHex(log.Data))
			case services.AbiEncodingMethod("AuctionSuccessful(uint256,uint256,address)"):
				err = successAuctionApostle(ctx, db, ec.Tx, chain, util.LogAnalysis(log.Data))
			case services.AbiEncodingMethod("NewBid(uint256,address,address,uint256,address,uint256,uint256)"):
				err = recordAuctionNewBid(ctx, db, ec.Tx, util.TrimHex(log.Topics[1]), chain, util.LogAnalysis(log.Data))
			}
		}
	}
	if err != nil {
		db.DbRollback()
		return err
	}

	if err := ec.NewUniqueTransaction(db, "clockAuctionApostle"); err != nil {
		return err
	}

	db.DbCommit()
	if db.Error != nil {
		return db.Error
	}

	if event == "successAuction" {
		log.Debug("get successAuction event. winner: %s tokenId: %s", winner, tokenId)
	}
	return err
}

func UnClaimedApostleList(ctx context.Context, district int, wallet []string) []string {
	list := myAuctionApostleList(ctx, wallet)
	var tokenIdArr []string
	for _, auction := range list {
		if auction.LastBidStart != 0 && time.Now().Unix()-int64(auction.LastBidStart) > apostleClaimTime(district) {
			tokenIdArr = append(tokenIdArr, auction.TokenId)
		}
	}
	return tokenIdArr
}

func FindOnsellApostle(ctx context.Context, seller, chain string, where []string) []AuctionApostle {
	db := util.WithContextDb(ctx)
	var aucs []AuctionApostle
	var query *gorm.DB
	if seller != "" {
		query = db.Where("seller = ? and status = ?", seller, AuctionGoing)
	} else {
		query = db.Where("status = ? and start_at <= ?", AuctionGoing, int(time.Now().Unix()))
	}
	query = query.Where("district = ?", GetDistrictByChain(chain))
	for _, w := range where {
		query = query.Where(w)
	}
	query.Find(&aucs)
	return aucs
}

func OnsellApostleList(ctx context.Context, seller string, chain string, where []string) (tokenIdArr []string, hasBidArr []string, priceMap map[string]decimal.Decimal, tokenMap map[string]*util.Token) {
	var aucs = FindOnsellApostle(ctx, seller, chain, where)
	priceMap = make(map[string]decimal.Decimal)
	tokenMap = make(map[string]*util.Token)
	if len(aucs) == 0 {
		return
	}
	for _, auction := range aucs {
		if auction.LastBidStart != 0 && time.Now().Unix()-int64(auction.LastBidStart) > apostleClaimTime(getNFTDistrict(auction.TokenId)) { // 排除unclaimed
			continue
		}
		priceMap[auction.TokenId] = auction.CurrentPrice()
		tokenMap[auction.TokenId] = util.Evo.GetToken(GetChainByDistrict(auction.District), auction.Currency)
		tokenIdArr = append(tokenIdArr, auction.TokenId)
		if auction.LastBidStart != 0 {
			hasBidArr = append(hasBidArr, auction.TokenId)
		}
	}
	return
}

func createAuctionApostle(ctx context.Context, db *util.GormDB, data []string, tx, chain string) error {
	tokenId := data[0]
	seller := util.AddHex(data[1][24:64], chain)
	apostle := GetApostleByTokenId(ctx, tokenId)
	if apostle == nil {
		_ = services.NewIssue("not create this Apostle :" + tx)
		return errors.New("not create this Apostle")
	}
	auc := AuctionApostle{ApostleId: apostle.ID, TokenId: apostle.TokenId, CreateTX: tx, Seller: seller}
	auc.Currency = util.AddHex(data[5][24:64], chain)

	tokenInfo := util.Evo.GetToken(chain, auc.Currency)

	auc.StartPrice = util.BigToDecimal(util.U256(data[2]), int32(tokenInfo.Decimals))
	auc.EndPrice = util.BigToDecimal(util.U256(data[3]), int32(tokenInfo.Decimals))
	auc.Duration = util.StringToInt(util.U256(data[4]).String())

	auc.StartAt = util.StringToInt(util.U256(data[6]).String())
	auc.District = apostle.District
	if err := auc.New(db); err != nil {
		return errors.New(" create this auction fail")
	}
	th := TransactionHistory{Tx: auc.CreateTX, BalanceAddress: auc.Seller, Action: TransactionHistoryApostleSale, TokenId: auc.TokenId, Chain: chain}
	_ = th.New(db)
	auctionAddress := util.GetContractAddress("clockAuctionApostle", chain)
	var originOwnerId uint
	originOwner := GetMemberByAddress(ctx, seller, GetChainByTokenId(tokenId))
	if originOwner != nil {
		originOwnerId = originOwner.ID
	}
	if err := apostle.TransferOwner(db, auctionAddress, apostleOnsell, seller, originOwnerId, chain); err != nil {
		db.DbRollback()
	}
	return nil
}

func recordAuctionNewBid(ctx context.Context, db *util.GormDB, tx, tokenId, chain string, logData []string) error {
	apostle := GetApostleByTokenId(ctx, tokenId)
	if apostle == nil {
		_ = services.NewIssue("not create this Apostle :" + tx)
		return errors.New("not create this Apostle")
	}
	auction := GetApostleAuction(ctx, apostle.TokenId, AuctionGoing)
	if auction == nil {
		return errors.New("auction not going")
	}
	previousBuyer := auction.LastBidder
	lastBuyer := util.AddHex(logData[0][24:64], chain)
	price := util.BigToDecimal(util.U256(logData[2]), util.GetTokenDecimals(chain))
	lastStart := util.StringToInt(util.U256(logData[4]).String())
	bidToken := util.AddHex(logData[3][24:64], chain)
	if price.GreaterThanOrEqual(auction.LastPrice) {
		db.Model(&auction).UpdateColumn(AuctionApostle{LastBidder: lastBuyer, LastPrice: price, LastBidStart: lastStart})
	}
	returnToLastBidder := util.BigToDecimal(util.U256(logData[5]), util.GetTokenDecimals(chain))
	ah := AuctionHistory{AuctionId: auction.ID, TokenId: tokenId, TxId: tx, Buyer: lastBuyer, BidPrice: price, StartAt: lastStart, BidToken: bidToken, AssetType: AssetApostle}
	if err := ah.New(ctx); err != nil {
		db.DbRollback()
		return err
	}
	if returnToLastBidder.Sign() > 0 { // 拍地退款
		th := TransactionHistory{Tx: tx, Chain: chain, BalanceAddress: previousBuyer, Action: TransactionHistoryApostleBidRefund, BalanceChange: returnToLastBidder, Currency: currencyRing, TokenId: tokenId}
		_ = th.New(db)
	}
	th := TransactionHistory{Tx: tx, Chain: chain, BalanceAddress: lastBuyer, Action: TransactionHistoryApostleBid, BalanceChange: price.Neg(), Currency: currencyRing, TokenId: tokenId}
	_ = th.New(db)
	return nil
}

func successAuctionApostle(ctx context.Context, db *util.GormDB, tx, chain string, logData []string) error {
	tokenId := util.TrimHex(logData[0])
	apostle := GetApostleByTokenId(ctx, tokenId)
	if apostle == nil {
		_ = services.NewIssue("not create this Apostle :" + tx)
		return errors.New("not create this Apostle")
	}
	auction := GetApostleAuction(ctx, apostle.TokenId, AuctionGoing)
	if auction == nil {
		return errors.New("auction not going")
	}
	price := util.BigToDecimal(util.U256(logData[1]), util.GetTokenDecimals(chain))
	winner := util.AddHex(logData[2][24:64], chain)
	db.Model(&auction).UpdateColumn(AuctionApostle{Winner: winner, FinalPrice: price, Status: AuctionFinish, ClaimTime: int(time.Now().Unix())})
	_ = apostle.TransferOwner(db, winner, apostleFresh, "", 0, chain)
	th := TransactionHistory{Tx: tx, Chain: chain, BalanceAddress: winner, Action: TransactionHistoryApostleSuccessAuction, TokenId: tokenId}
	_ = th.New(db)
	return nil
}

func cancelAuctionApostle(ctx context.Context, db *util.GormDB, tx, tokenId string) error {
	if auction := GetApostleAuction(ctx, tokenId, AuctionGoing); auction == nil {
		return errors.New("not find")
	} else {
		apostle := GetApostleByTokenId(ctx, tokenId)
		db.Model(&auction).UpdateColumn("status", AuctionCancel)
		_ = apostle.TransferOwner(db, auction.Seller, apostleFresh, "", 0, GetChainByTokenId(tokenId))
		th := TransactionHistory{Tx: tx, Chain: GetChainByTokenId(tokenId), BalanceAddress: auction.Seller, Action: TransactionHistoryApostleCancelSale, TokenId: tokenId}
		_ = th.New(db)
	}
	return nil
}

func GetApostleAuction(ctx context.Context, tokenId, status string) *AuctionApostle {
	db := util.WithContextDb(ctx)
	var auc AuctionApostle
	query := db.Where("token_id = ? and status = ?", tokenId, status).First(&auc)
	if query.Error != nil || query.RecordNotFound() {
		return nil
	}
	return &auc
}

func getAuctionApostleReversalKey(ctx context.Context, ids []uint, status string) map[uint]AuctionApostle {
	if len(ids) < 1 {
		return nil
	}
	db := util.WithContextDb(ctx)
	var auctions []AuctionApostle
	var query *gorm.DB
	if len(ids) == 1 {
		query = db.Where("status = ?", status).Where("apostle_id = ?", ids[0]).Find(&auctions)
	} else {
		query = db.Where("status = ?", status).Where("apostle_id in (?)", ids).Find(&auctions)
	}
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	results := make(map[uint]AuctionApostle)
	for _, v := range auctions {
		results[v.ApostleId] = v
	}
	return results
}

func myAuctionApostleList(ctx context.Context, wallet []string) []AuctionApostle {
	db := util.WithContextDb(ctx)
	var aucs []AuctionApostle
	var query *gorm.DB
	if len(wallet) < 1 {
		query = db.Where("last_bidder != ? and status = ?", noneAddress, AuctionGoing).Find(&aucs)
	} else {
		query = db.Where("last_bidder = ? and status = ?", wallet[0], AuctionGoing).Find(&aucs)
	}
	if query.RecordNotFound() {
		return aucs
	}
	return aucs
}

func filterCanSell(ctx context.Context, apostles []ApostleJson, chain string) []ApostleJson {
	var newApostles []ApostleJson
	tokenIds, _, _, _ := OnsellApostleList(ctx, "", chain, nil) // 排除未领取的
	for _, v := range apostles {
		if util.StringInSlice(v.TokenId, tokenIds) {
			newApostles = append(newApostles, v)
		}
	}
	return newApostles
}
