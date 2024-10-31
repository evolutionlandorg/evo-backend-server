package models

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
)

type ApostleReward struct {
	gorm.Model
	TokenId   string `json:"token_id"`
	Wallet    string `json:"wallet"`
	Tx        string `sql:"default: null" json:"tx"`
	ExpiredAt int    `json:"expired_at"`
	Reason    string `json:"reason"`
}

func (ec *EthTransactionCallback) TakeBackNFTCallback(ctx context.Context) (err error) {
	if getTransactionDeal(ctx, ec.Tx, "takeBackNFT") != nil {
		return errors.New("tx exist")
	}
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("takeBackNFT", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			case services.AbiEncodingMethod("TakenBackNFT(address,uint256,uint256)"):
				address := util.AddHex(util.TrimHex(log.Topics[1])[24:64], chain)
				logSlice := util.LogAnalysis(log.Data)
				confirmApostleReward(db, address, ec.Tx, logSlice[0])
			}
		}
	}

	if err := ec.NewUniqueTransaction(db, "takeBackNFT"); err != nil {
		return err
	}
	db.DbCommit()
	return db.Error
}

//func InitApostleAirDropRewardTokenId() {
//	for index := 101; index <= 1100; index++ {
//		locationIndex := fmt.Sprintf("%032s", fmt.Sprintf("%x", index))
//		tokenId := fmt.Sprintf("2a010001010001020000000000000001%s", locationIndex)
//		txn := util.DbBegin(context.TODO())
//		txn.FirstOrCreate(&ApostleReward{}, map[string]interface{}{"token_id": tokenId})
//		txn.DbCommit()
//	}
//}

func confirmApostleReward(db *util.GormDB, address, tx, tokenId string) {
	db.Model(&ApostleReward{}).Where(ApostleReward{Wallet: address, TokenId: tokenId}).UpdateColumn("tx", tx)
}

func UnClaimApostleReward(ctx context.Context, wallet string) []string {
	var tokenId []string
	db := util.WithContextDb(ctx)
	db.Model(&ApostleReward{}).Where("wallet = ? and expired_at >= ? and tx is null", wallet, int(time.Now().Unix())).Pluck("token_id", &tokenId)
	return tokenId
}

//func AddApostleReward(txn *util.GormDB, wallet, reason string) string {
//	expiredAt := int(time.Now().Add(time.Hour * 24 * 365).Unix())
//	var exist ApostleReward
//	query := txn.Model(&ApostleReward{}).Where("wallet = ? and reason = ?", wallet, reason).First(&exist)
//	if query.RecordNotFound() {
//		txn.Model(&ApostleReward{}).Where("wallet = ?", "").Limit(1).UpdateColumn(ApostleReward{Wallet: wallet, ExpiredAt: expiredAt, Reason: reason})
//		return ""
//	} else {
//		return wallet
//	}
//}

func CanTakeNft(ctx context.Context, wallet, tokenId string) *ApostleReward {
	var reward ApostleReward
	db := util.WithContextDb(ctx)
	query := db.Model(&ApostleReward{}).Where("token_id = ? and wallet = ? and expired_at >= ? and tx is null", tokenId, wallet, int(time.Now().Unix())).First(&reward)
	if query.Error != nil || query.RecordNotFound() {
		return nil
	}
	return &reward
}
