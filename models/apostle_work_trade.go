package models

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type ApostleWorkTrade struct {
	gorm.Model
	ApostleId uint            `json:"apostle_id"`
	TokenId   string          `json:"token_id"`
	CreateTX  string          `json:"create_tx"`
	Seller    string          `json:"seller"`
	Price     decimal.Decimal `json:"price" sql:"type:decimal(36,18);"`
	Duration  int             `json:"duration"`
	StartAt   int             `json:"start_at"`
	Status    string          `json:"status"` // going拍卖中/finish已完成/cancel取消/租赁完成 over
	Winner    string          `json:"winner"`
	Activity  string          `json:"activity"`
	District  int             `json:"district"`
	Currency  string          `json:"currency"`
}

type ApostleWorkInfo struct {
	Seller     string          `json:"seller"`
	Owner      string          `json:"owner"`
	SellerName string          `json:"seller_name"`
	Price      decimal.Decimal `json:"price" sql:"type:decimal(36,18);"`
	Duration   int             `json:"duration"`
	DigElement string          `json:"dig_element"`
	Lon        int             `json:"lon"`
	Lat        int             `json:"lat"`
	Strength   decimal.Decimal `json:"strength" sql:"type:decimal(36,18);"`
	LandId     uint            `json:"land_id"`
	TokenIndex int             `json:"token_index"`
}

func (awt *ApostleWorkTrade) New(ctx context.Context, _ *util.GormDB) error {
	result := util.WithContextDb(ctx).Create(&awt)
	return result.Error
}

func (ap *Apostle) ApostleRentInfo(ctx context.Context) *ApostleWorkInfo {
	if ap.Status == apostleHiring || ap.Status == apostleWorking {
		db := util.WithContextDb(ctx)
		var aw ApostleWorkInfo
		db.Table("land_apostles").Select("lon,lat,land_id,strength,dig_element,lands.token_index,owner").
			Joins("join lands on land_apostles.land_id = lands.id").
			Where("apostle_id =?", ap.ID).Scan(&aw)
		aw.LandId = uint(aw.TokenIndex + 1)
		auction := GetApostleWork(ctx, ap.TokenId, AuctionFinish)
		if auction == nil {
			return &aw
		}
		seller := GetMemberByAddress(ctx, auction.Seller, GetChainByTokenId(ap.TokenId))
		if seller != nil {
			aw.SellerName = seller.Name
		}
		aw.Seller = auction.Seller
		aw.Price = auction.Price
		aw.Duration = auction.Duration
		return &aw
	}
	return nil
}

func (ec *EthTransactionCallback) TokenUseCallback(ctx context.Context) (err error) {

	if getTransactionDeal(ctx, ec.Tx, "tokenUse") != nil {
		return errors.New("tx exist")
	}
	chain := ec.Receipt.ChainSource
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("tokenUse", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			case services.AbiEncodingMethod("OfferCreated(uint256,uint256,uint256,address,address)"):
				dataSlice := util.LogAnalysis(log.Data)
				err = offerCreated(ctx, db, ec.Tx, chain, util.TrimHex(log.Topics[1]), dataSlice)
			case services.AbiEncodingMethod("OfferCancelled(uint256)"):
				err = offerCancelled(ctx, db, ec.Tx, chain, util.TrimHex(log.Data))
			case services.AbiEncodingMethod("OfferTaken(uint256,address,address,uint256,uint256)"):
				dataSlice := util.LogAnalysis(log.Data)
				err = offerTaken(ctx, db, ec.Tx, chain, util.TrimHex(log.Topics[1]), dataSlice)
			case services.AbiEncodingMethod("TokenUseRemoved(uint256,address,address,address)"): // 打工时间到期
				err = tokenUseRemoved(ctx, db, ec.Tx, chain, util.TrimHex(log.Topics[1]))
			}
		}
	}
	if err != nil {
		return err
	}

	if err := ec.NewUniqueTransaction(db, "tokenUse"); err != nil {
		return err
	}

	db.DbCommit()
	return db.Error
}

func getApostleWorkTradeReversalKey(ctx context.Context, ids []uint, status string) map[uint]ApostleWorkTrade {
	if len(ids) < 1 {
		return nil
	}
	db := util.WithContextDb(ctx)
	var auctions []ApostleWorkTrade
	query := db.Where("status = ?", status).Where("apostle_id in (?)", ids).Find(&auctions)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	results := make(map[uint]ApostleWorkTrade)
	for _, v := range auctions {
		results[v.ApostleId] = v
	}
	return results
}

func GetApostleWork(ctx context.Context, tokenId, status string) *ApostleWorkTrade {
	db := util.WithContextDb(ctx)
	var auc ApostleWorkTrade
	query := db.Where("token_id = ? and status = ?", tokenId, status).First(&auc)
	if query.Error != nil || query.RecordNotFound() {
		return nil
	}
	return &auc
}

func offerCreated(ctx context.Context, db *util.GormDB, tx, chain, tokenId string, data []string) error {
	apostle := GetApostleByTokenId(ctx, tokenId)
	if apostle == nil {
		_ = services.NewIssue("not create this Apostle :" + tx)
		return errors.New("not create this Apostle")
	}
	aw := ApostleWorkTrade{ApostleId: apostle.ID, TokenId: apostle.TokenId, CreateTX: tx, Status: AuctionGoing}
	aw.Duration = util.StringToInt(util.U256(data[0]).String())
	aw.Price = util.BigToDecimal(util.U256(data[1]), util.GetTokenDecimals(chain))
	aw.Activity = util.AddHex(data[2][24:64], chain)
	aw.Seller = util.AddHex(data[3][24:64], chain)
	aw.Currency = util.GetContractAddress("ring", chain)

	aw.District = apostle.District
	if err := aw.New(ctx, db); err != nil {
		return errors.New(" create this auction fail")
	}
	th := TransactionHistory{Tx: aw.CreateTX, Chain: chain, BalanceAddress: aw.Seller, Action: TransactionHistoryApostleRent, TokenId: aw.TokenId}
	_ = th.New(db)
	auctionAddress := util.GetContractAddress("tokenUse", chain)
	var originOwnerId uint
	originOwner := GetMemberByAddress(ctx, aw.Seller, GetChainByTokenId(tokenId))
	if originOwner != nil {
		originOwnerId = originOwner.ID
	}
	if err := apostle.TransferOwner(db, auctionAddress, apostleRent, aw.Seller, originOwnerId, chain); err != nil {
		db.DbRollback()
	}
	return nil
}

func offerCancelled(ctx context.Context, db *util.GormDB, tx, chain, tokenId string) error {
	if auction := GetApostleWork(ctx, tokenId, AuctionGoing); auction == nil {
		return errors.New("not find")
	} else {
		apostle := GetApostleByTokenId(ctx, tokenId)
		db.Model(&auction).UpdateColumn("status", AuctionCancel)
		_ = apostle.TransferOwner(db, auction.Seller, apostleFresh, "", 0, chain)
		th := TransactionHistory{Tx: tx, Chain: chain, BalanceAddress: auction.Seller, Action: TransactionHistoryApostleRentCancel, TokenId: tokenId}
		_ = th.New(db)
	}
	return nil
}

func offerTaken(ctx context.Context, db *util.GormDB, tx, chain, tokenId string, data []string) error {
	if auction := GetApostleWork(ctx, tokenId, AuctionGoing); auction == nil {
		return errors.New("not find")
	} else {
		apostle := GetApostleByTokenId(ctx, tokenId)
		winner := util.AddHex(data[0][24:64], chain)
		startAt := util.StringToInt(util.U256(data[2]).String())
		db.Model(&auction).UpdateColumn(ApostleWorkTrade{Status: AuctionFinish, Winner: winner, StartAt: startAt})
		_ = apostle.TransferOwner(db, winner, apostleHiring, auction.Seller, apostle.OriginId, chain)
		th := TransactionHistory{Tx: tx, Chain: chain, BalanceAddress: winner, BalanceChange: auction.Price.Neg(), Action: TransactionHistoryApostleRentSuccess, TokenId: tokenId, Currency: currencyRing}
		_ = th.New(db)
		th = TransactionHistory{Tx: tx, Chain: chain, BalanceAddress: auction.Seller, BalanceChange: auction.Price.Mul(decimal.RequireFromString(apostleTradeTex)), Action: TransactionHistoryApostleRentSuccess, TokenId: tokenId, Currency: currencyRing}
		_ = th.New(db)
	}
	return nil
}

func tokenUseRemoved(ctx context.Context, db *util.GormDB, tx, chain, tokenId string) error {
	if auction := GetApostleWork(ctx, tokenId, AuctionFinish); auction == nil {
		return errors.New("not find")
	} else {
		db.Model(&auction).UpdateColumn("status", AuctionOver)
		apostle := GetApostleByTokenId(ctx, tokenId)
		_ = apostle.TransferOwner(db, auction.Seller, apostleFresh, "", 0, chain)
		th := TransactionHistory{Tx: tx, Chain: chain, BalanceAddress: auction.Seller, Action: TransactionHistoryApostleWorkingStop, TokenId: tokenId}
		_ = th.New(db)
	}
	return nil
}

func filterCanWorking(ctx context.Context, owner string, apostles []ApostleJson) []ApostleJson {
	db := util.WithContextDb(ctx)
	var newApostles []ApostleJson
	var babyMother []string
	db.Table("apostles").Where("owner=?", owner).Where("status=?", apostleBirth).Pluck("mother", &babyMother)
	for _, v := range apostles {
		if util.StringInSlice(v.TokenId, babyMother) { // birthUnclaimed
			continue
		}
		if (v.Status == "fresh" && v.ColdDownEnd < int(time.Now().Unix())) || v.Status == "hiring" {
			newApostles = append(newApostles, v)
		}
	}
	return newApostles
}

func FindWorkerPrice(ctx context.Context, chain string, where ...string) []ApostleWorkTrade {
	db := util.WithContextDb(ctx)
	var aucs []ApostleWorkTrade
	query := db.Where("status = ?", AuctionGoing).Where("district = ?", GetDistrictByChain(chain))
	for _, w := range where {
		query = query.Where(w)
	}
	query.Find(&aucs)
	return aucs
}

func WorkerPriceList(ctx context.Context, chain string, where ...string) (priceMap map[string]decimal.Decimal, tokenMap map[string]*util.Token) {
	var aucs = FindWorkerPrice(ctx, chain, where...)
	priceMap = make(map[string]decimal.Decimal)
	tokenMap = make(map[string]*util.Token)
	if len(aucs) == 0 {
		return
	}
	for _, auction := range aucs {
		priceMap[auction.TokenId] = auction.Price.Div(decimal.New(int64(auction.Duration/86400), 0))
		tokenMap[auction.TokenId] = util.Evo.GetToken(GetChainByDistrict(auction.District), auction.Currency)
	}
	return
}
