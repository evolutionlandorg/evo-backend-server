package models

import (
	"context"
	"time"

	"github.com/evolutionlandorg/evo-backend/util"
)

type UniqueTransaction struct {
	ID          uint `gorm:"primary_key" json:"id"`
	CreatedAt   time.Time
	Action      string `json:"action"`
	Tx          string `json:"tx"`
	Chain       string `json:"chain"`
	BlockNum    string `json:"block_num"`
	Confirm     bool   `json:"confirm" sql:"default: 1;"`
	Unprocessed bool   `json:"unprocessed"`
}

func (ec *EthTransactionCallback) NewUniqueTransaction(txn *util.GormDB, action string) error {
	ut := UniqueTransaction{Tx: ec.Tx, Action: action, Chain: ec.Receipt.ChainSource, BlockNum: util.U256(ec.Receipt.BlockNumber).String()}
	if ut.Chain == "Eth" { // examine any eth transaction
		ut.Confirm = false
	}
	result := txn.Create(&ut)
	return result.Error
}

func getTransactionDeal(ctx context.Context, tx, action string) *UniqueTransaction {
	db := util.WithContextDb(ctx)
	var ut UniqueTransaction
	query := db.Where("tx = ? and action = ?", tx, action).First(&ut)
	if query.Error != nil || query == nil || query.RecordNotFound() {

		return nil
	}
	return &ut
}
