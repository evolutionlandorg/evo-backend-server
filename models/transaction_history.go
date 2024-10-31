package models

import (
	"context"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/shopspring/decimal"
)

type TransactionHistory struct {
	ID             uint `json:"id"`
	CreatedAt      time.Time
	Tx             string          `json:"tx"`
	BalanceChange  decimal.Decimal `json:"balance_change" sql:"type:decimal(32,16);"`
	BalanceAddress string          `json:"balance_address"`
	TokenId        string          `json:"token_id"`
	Action         string          `json:"action"`
	Currency       string          `json:"currency"`
	AddTime        int             `json:"add_time"`
	Coordinate     string          `json:"coordinate"`
	Extra          string          `json:"extra"`
	Chain          string          `json:"chain"`
}

type TransactionHistoryJson struct {
	Tx             string          `json:"tx"`
	Action         string          `json:"action"`
	TokenId        string          `json:"token_id"`
	BalanceChange  decimal.Decimal `json:"balance_change"`
	BalanceAddress string          `json:"balance_address"`
	Currency       string          `json:"currency"`
	AddTime        int             `json:"add_time"`
	Coordinate     string          `json:"coordinate"`
	Extra          string          `json:"extra"`
}

type EthTransactionQuery struct {
	WhereQuery struct {
		BalanceAddress string
		Action         string
		Chain          string
	}
	Row  int
	Page int
}

func (th *TransactionHistory) New(db *util.GormDB) error {
	th.ID = 0
	th.AddTime = int(time.Now().Unix())
	result := db.Create(&th)
	return result.Error
}

func (etq *EthTransactionQuery) GetTransactionHistory(ctx context.Context) (*[]TransactionHistoryJson, int) {
	db := util.WithContextDb(ctx)
	var (
		ethTran []TransactionHistoryJson
		count   int
	)
	query := db.Table("transaction_histories").Where(etq.WhereQuery).
		Offset(etq.Page * etq.Row).Limit(etq.Row).Order("id desc").
		Scan(&ethTran)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil, 0
	}
	db.Table("transaction_histories").Where(etq.WhereQuery).Count(&count)
	return &ethTran, count
}

func setTransactionHistoryExtra(txn *util.GormDB, tx, action, extra string) {
	txn.Model(&TransactionHistory{}).Where("tx =?", tx).Where("action=?", action).UpdateColumn(TransactionHistory{Extra: extra})
}

func GetTransCount24H(ctx context.Context, chain string) uint64 {
	var count uint64
	util.WithContextDb(ctx).Model(&TransactionHistory{}).Where("chain = ?", chain).Where("add_time >= ?", time.Now().UTC().Unix()-86400).Count(&count)
	return count
}

func GetTransCount7D(ctx context.Context, chain string) uint64 {
	var count uint64
	util.WithContextDb(ctx).Model(&TransactionHistory{}).Where("chain = ?", chain).Where("add_time >= ?", time.Now().UTC().Unix()-7*86400).Count(&count)
	return count
}

func GetTransAmount24H(ctx context.Context, chain, ring string) decimal.Decimal {
	var (
		trans []TransactionHistory
		sum   decimal.Decimal
	)
	util.WithContextDb(ctx).Model(&TransactionHistory{}).Where("chain = ?", chain).Where("add_time >= ?", time.Now().UTC().Unix()-86400).Find(&trans)
	for _, v := range trans {
		if strings.EqualFold(v.Currency, ring) {
			sum = sum.Add(v.BalanceChange.Abs())
		}
	}
	return sum
}

func GetTransAmount7D(ctx context.Context, chain, ring string) decimal.Decimal {
	var (
		trans []TransactionHistory
		sum   decimal.Decimal
	)
	util.WithContextDb(ctx).Model(&TransactionHistory{}).Where("chain = ?", chain).Where("add_time >= ?", time.Now().UTC().Unix()-7*86400).Find(&trans)
	for _, v := range trans {
		if strings.EqualFold(v.Currency, ring) {
			sum = sum.Add(v.BalanceChange.Abs())
		}
	}
	return sum
}

func GetTVL(ctx context.Context, chain, ring string) decimal.Decimal {
	var (
		trans []TransactionHistory
		sum   decimal.Decimal
	)
	util.WithContextDb(ctx).Model(&TransactionHistory{}).Where("chain = ?", chain).Find(&trans)
	for _, v := range trans {
		if strings.EqualFold(v.Currency, ring) {
			sum = sum.Add(v.BalanceChange.Abs())
		}
	}
	return sum
}
