package models

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type ApostleFertility struct {
	gorm.Model
	ApostleId  uint            `json:"apostle_id"`
	TokenId    string          `json:"token_id"`
	CreateTX   string          `json:"create_tx"`
	Seller     string          `json:"seller"`
	StartPrice decimal.Decimal `json:"start_price" sql:"type:decimal(36,18);"`
	EndPrice   decimal.Decimal `json:"end_price" sql:"type:decimal(36,18);"`
	Duration   int             `json:"duration"`
	StartAt    int             `json:"start_at"`
	Status     string          `json:"status"` // going/finish/cancel
	Winner     string          `json:"winner"`
	FinalPrice decimal.Decimal `json:"final_price" sql:"type:decimal(36,18);"`
	District   int             `json:"district"`
	Currency   string          `json:"currency"`
}

var apostleBirthFee = decimal.RequireFromString(fertilityPrice)

func (af *ApostleFertility) New(db *util.GormDB) error {
	result := db.Create(&af)
	return result.Error
}

func (af *ApostleFertility) CurrentPrice() decimal.Decimal {
	now := int(time.Now().Unix())
	if af == nil || af.Status != AuctionGoing || now < af.StartAt || af.Duration == 0 {
		return decimal.RequireFromString("-1")
	}
	secondsPassed := now - af.StartAt

	if secondsPassed > af.Duration {
		return af.EndPrice
	} else {
		totalPriceInRINGChange := af.EndPrice.Sub(af.StartPrice)
		secondsDecimal := decimal.RequireFromString(util.IntToString(secondsPassed))
		durationDecimal := decimal.RequireFromString(util.IntToString(af.Duration))
		currentPriceInRINGChange := totalPriceInRINGChange.Mul(secondsDecimal.Div(durationDecimal))
		return af.StartPrice.Add(currentPriceInRINGChange).Round(18)
	}
}

//func getApostleFertilityReversalKey(ids []uint, status string) map[uint]ApostleFertility {
//	if len(ids) < 1 {
//		return nil
//	}
//	db := util.DB
//	var auctions []ApostleFertility
//	query := db.Where("status = ?", status).Where("apostle_id in (?)", ids).Find(&auctions)
//	if query.Error != nil || query == nil || query.RecordNotFound() {
//		return nil
//	}
//	results := make(map[uint]ApostleFertility)
//	for _, v := range auctions {
//		results[v.ApostleId] = v
//	}
//	return results
//}

func GetApostleFertility(ctx context.Context, tokenId, status string) *ApostleFertility {
	db := util.WithContextDb(ctx)
	var auc ApostleFertility
	query := db.Where("token_id = ? and status = ?", tokenId, status).First(&auc)
	if query.Error != nil || query.RecordNotFound() {
		return nil
	}
	return &auc
}

// ApostleFertilityCallback 使徒生育
func (ec *EthTransactionCallback) ApostleFertilityCallback(ctx context.Context) (err error) {
	if getTransactionDeal(ctx, ec.Tx, "ApostleFertility") != nil {
		return errors.New("tx exist")
	}
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	var event, winner, tokenId string
	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("apostleFertility", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			case services.AbiEncodingMethod("AuctionCreated(uint256,address,uint256,uint256,uint256,address,uint256)"):
				dataSlice := util.LogAnalysis(log.Data)
				err = createApostleFertility(ctx, db, dataSlice, ec.Tx, chain)
			case services.AbiEncodingMethod("AuctionCancelled(uint256)"):
				err = cancelApostleFertility(ctx, db, ec.Tx, util.TrimHex(log.Data))
			case services.AbiEncodingMethod("AuctionSuccessful(uint256,uint256,address)"):
				dataSlice := util.LogAnalysis(log.Data)
				err = finishApostleFertility(ctx, db, ec.Tx, chain, dataSlice)
			}
		}
	}
	if err != nil {
		db.DbRollback()
		return err
	}

	if err := ec.NewUniqueTransaction(db, "ApostleFertility"); err != nil {
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

func createApostleFertility(ctx context.Context, db *util.GormDB, data []string, tx, chain string) error {
	tokenId := data[0]
	seller := util.AddHex(data[1][24:64], chain)
	apostle := GetApostleByTokenId(ctx, tokenId)
	if apostle == nil {
		_ = services.NewIssue("not create this Apostle :" + tx)
		return errors.New("not create this Apostle")
	}
	auc := ApostleFertility{ApostleId: apostle.ID, TokenId: apostle.TokenId, CreateTX: tx, Seller: seller, Status: AuctionGoing}
	auc.StartPrice = util.BigToDecimal(util.U256(data[2]), util.GetTokenDecimals(chain))
	auc.EndPrice = util.BigToDecimal(util.U256(data[3]), util.GetTokenDecimals(chain))
	auc.Duration = util.StringToInt(util.U256(data[4]).String())
	auc.StartAt = util.StringToInt(util.U256(data[6]).String())
	auc.Currency = util.ToChecksumAddress(data[5][24:64])
	auc.District = apostle.District
	if err := auc.New(db); err != nil {
		return errors.New(" create this auction fail")
	}
	th := TransactionHistory{Tx: auc.CreateTX, Chain: chain, BalanceAddress: auc.Seller, Action: TransactionHistoryApostleFertility, TokenId: auc.TokenId}
	_ = th.New(db)
	auctionAddress := util.GetContractAddress("apostleFertility", chain)
	var originOwnerId uint
	originOwner := GetMemberByAddress(ctx, seller, GetChainByTokenId(tokenId))
	if originOwner != nil {
		originOwnerId = originOwner.ID
	}
	if err := apostle.TransferOwner(db, auctionAddress, apostleFertility, seller, originOwnerId, chain); err != nil {
		db.DbRollback()
	}
	return nil
}

func cancelApostleFertility(ctx context.Context, db *util.GormDB, tx, tokenId string) error {
	if auction := GetApostleFertility(ctx, tokenId, AuctionGoing); auction == nil {
		return errors.New("not find")
	} else {
		apostle := GetApostleByTokenId(ctx, tokenId)
		db.Model(&auction).UpdateColumn("status", AuctionCancel)
		_ = apostle.TransferOwner(db, auction.Seller, apostleFresh, "", 0, GetChainByTokenId(tokenId))
		th := TransactionHistory{Tx: tx, Chain: GetChainByTokenId(tokenId), BalanceAddress: auction.Seller, Action: TransactionHistoryApostleFertilityCancel, TokenId: tokenId}
		_ = th.New(db)
	}
	return nil
}

func finishApostleFertility(ctx context.Context, db *util.GormDB, tx, chain string, logData []string) error {
	tokenId := util.TrimHex(logData[0])
	price := util.BigToDecimal(util.U256(logData[1]), util.GetTokenDecimals(chain))
	winner := util.AddHex(logData[2][24:64], chain)
	apostle := GetApostleByTokenId(ctx, tokenId)
	if apostle == nil {
		_ = services.NewIssue("not create this Apostle :" + tx)
		return errors.New("not create this Apostle")
	}
	auction := GetApostleFertility(ctx, apostle.TokenId, AuctionGoing)
	if auction == nil {
		return errors.New("auction not going")
	}
	db.Model(&auction).UpdateColumn(ApostleFertility{Winner: winner, FinalPrice: price, Status: AuctionFinish})
	_ = apostle.TransferOwner(db, auction.Seller, apostleFresh, "", 0, GetChainByTokenId(tokenId))
	th := TransactionHistory{Tx: tx, Chain: chain, BalanceAddress: winner, BalanceChange: price.Neg(), Action: TransactionHistoryApostleFertilitySuccess, TokenId: tokenId, Currency: currencyRing}
	_ = th.New(db)
	th = TransactionHistory{Tx: tx, Chain: chain, BalanceAddress: auction.Seller, BalanceChange: price.Mul(decimal.RequireFromString(apostleTradeTex)), Action: TransactionHistoryApostleFertilitySuccess, TokenId: tokenId, Currency: currencyRing}
	_ = th.New(db)
	return nil
}

func CanSiringApostle(ctx context.Context, address, sireTokenId string) []string {
	var ids []string
	var apostles []Apostle
	var babyMother []string
	db := util.WithContextDb(ctx)
	if sireTokenId == "" {
		return ids
	}
	sire := GetApostleByTokenId(ctx, sireTokenId)
	if sire == nil {
		return ids
	}
	db.Table("apostles").Where("owner=?", address).Where("status=?", apostleFresh).Where("cold_down_end <?", time.Now().Unix()).Find(&apostles)
	db.Table("apostles").Where("owner=?", address).Where("status=?", apostleBirth).Pluck("mother", &babyMother)
	for _, v := range apostles {
		if util.StringInSlice(v.TokenId, babyMother) { // birthUnclaimed
			continue
		}
		if sire.Gender != v.Gender { // 性别不同
			if getApostlesAlienFromGenes(sire.Genes) == getApostlesAlienFromGenes(v.Genes) { // 种族相同
				if sire.Mother != "" { // 非0代
					if sire.Mother != v.Mother && sire.Mother != v.Father && sire.Father != v.Mother && sire.Father != v.Father && v.Mother != sire.TokenId && v.Father != sire.TokenId && sire.Mother != v.TokenId && sire.Father != v.TokenId {
						ids = append(ids, v.TokenId)
					}
				} else { // 0代
					if v.Mother != sire.TokenId && v.Father != sire.TokenId { // 不能跟自己的孩子生
						ids = append(ids, v.TokenId)
					}
				}

			}
		}
	}
	return ids
}

func FindSiringApostlePrice(ctx context.Context, chain string, where ...string) []ApostleFertility {
	db := util.WithContextDb(ctx)
	var aucs []ApostleFertility
	query := db.Where("status = ?", AuctionGoing).Where("district = ?", GetDistrictByChain(chain))
	for _, w := range where {
		query = query.Where(w)
	}
	query.Find(&aucs)
	return aucs
}

func SiringApostlePriceList(ctx context.Context, chain string, where ...string) (priceMap map[string]decimal.Decimal, tokenMap map[string]*util.Token) {
	aucs := FindSiringApostlePrice(ctx, chain, where...)
	priceMap = make(map[string]decimal.Decimal)
	tokenMap = make(map[string]*util.Token)
	if len(aucs) == 0 {
		return
	}
	for _, auction := range aucs {
		priceMap[auction.TokenId] = auction.CurrentPrice()
		tokenMap[auction.TokenId] = util.Evo.GetToken(GetChainByDistrict(auction.District), auction.Currency)
	}
	return
}
