package models

import (
	"context"
	"time"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
)

type EthTransaction struct {
	gorm.Model
	From        string `json:"from"`
	To          string `json:"to"`
	Tx          string `json:"tx"`
	ChainUrl    string `json:"chain_url"`
	ReceiptsLog string `json:"receipts_log" sql:"type:text;"`
	Status      string `json:"status"`
	MemberId    uint   `json:"member_id"`
	Genre       string `json:"genre"` // sale transfer withdraw
	TokenId     string `json:"token_id"`
	Action      string `json:"action"`
	Chain       string `json:"chain"`
}

func (et *EthTransaction) UpdateEthTransaction(ctx context.Context, tx string) {
	db := util.WithContextDb(ctx)
	db.Model(et).Where("tx = ? ", tx).Updates(et)
	db.Table("transaction_histories").Where("tx=? and action=?", tx, TransactionHistoryPending).Delete(&TransactionHistory{})
}

func CurrentPendingLand(ctx context.Context, wallet string) map[string]string {
	db := util.WithContextDb(ctx)
	var ethTrans []EthTransaction
	pendingLands := make(map[string]string)
	db.Table("eth_transactions").Where("`from` = ?", wallet).Where("status = ?", "Pending").Where("token_id !=?", "").Find(&ethTrans)
	for _, v := range ethTrans {
		pendingLands[v.TokenId] = v.Tx
	}
	return pendingLands
}

func GetEthTransactionPending(ctx context.Context) []EthTransaction {
	db := util.WithContextDb(ctx)
	var ethTran []EthTransaction
	query := db.Where("status = ?", "Pending").Where("created_at >?", time.Now().AddDate(0, 0, -3)).Find(&ethTran)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	return ethTran
}
