package models

import (
	"context"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type Withdraw struct {
	gorm.Model
	TxId      string          `json:"txid"`
	Currency  string          `json:"currency"`
	Status    string          `json:"status"`
	MemberId  uint            `json:"member_id"`
	AccountId uint            `json:"account_id"`
	Amount    decimal.Decimal `json:"amount" sql:"type:decimal(32,16);"`
	SignNonce string          `json:"sign_nonce"`
}

// NewWithdraw params currency to _
func (a *Account) NewWithdraw(db *util.GormDB, signNonce, tx string, amount decimal.Decimal, _ string) error {
	var withdraw = Withdraw{Currency: a.Currency, Status: "done", MemberId: a.MemberId, AccountId: a.ID, Amount: amount, SignNonce: signNonce, TxId: tx}
	result := db.Create(&withdraw)
	return result.Error
}

// TakeBackCallback
func (ec *EthTransactionCallback) TakeBackCallback(ctx context.Context) error {

	if getTransactionDeal(ctx, ec.Tx, "withdraw") != nil {
		return errors.New("tx exist")
	}
	chain := ec.Receipt.ChainSource
	var (
		err error
	)

	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("takeBack", chain)) {
			if util.AddHex(log.Topics[0]) == services.AbiEncodingMethod("TakedBack(address,uint256,uint256)") {
				wallet := util.AddHex(util.TrimHex(log.Topics[1])[24:64], chain)
				nonce := util.StringToInt(util.U256(log.Topics[2]).String())
				member := GetMemberByAddress(ctx, wallet, chain)
				if member == nil {
					return errors.New("record not find")
				}

				db := util.DbBegin(ctx)

				amount := util.BigToDecimal(util.U256(log.Data), util.GetTokenDecimals(chain))
				account := member.TouchAccount(ctx, currencyRing, wallet, chain)
				if err := account.subBalance(db, amount, "withdraw"); err != nil {
					db.DbRollback()
					return errors.Wrap(err, "balance insufficient")
				}
				err = member.updateWithdrawNonce(ctx, db, nonce+1, chain, currencyRing)
				if err != nil {
					db.DbRollback()
					return errors.Wrap(err, "new withdraw error")
				}
				th := TransactionHistory{Tx: ec.Tx, Chain: chain, BalanceAddress: wallet, BalanceChange: amount, Action: TransactionHistoryWithdraw, Currency: currencyRing}
				_ = th.New(db)
				if err = account.NewWithdraw(db, util.IntToString(nonce), ec.Tx, amount, currencyRing); err != nil {
					db.DbRollback()
					return errors.Wrap(err, "new withdraw error")
				}
				db.DbCommit()
				if db.Error != nil {
					db.DbRollback()
					return errors.Wrap(err, "commit TakeBackCallback error")
				}
				go afterWithdraw(chain)
			}
		}
	}

	unIrDb := util.DbBegin(ctx)
	defer unIrDb.Rollback()
	if err := ec.NewUniqueTransaction(unIrDb, "withdraw"); err != nil {
		return errors.Wrap(err, "create unique transaction error")
	}
	unIrDb.DbCommit()
	return unIrDb.Error
}

// TakeBackKtonCallback
func (ec *EthTransactionCallback) TakeBackKtonCallback(ctx context.Context) error {

	if getTransactionDeal(ctx, ec.Tx, "TakeBackKton") != nil {
		return errors.New("tx exist")
	}
	chain := ec.Receipt.ChainSource
	var err error
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("TakeBackKton", chain)) {
			if util.AddHex(log.Topics[0]) == services.AbiEncodingMethod("TakedBack(address,uint256,uint256)") {
				wallet := util.AddHex(util.TrimHex(log.Topics[1])[24:64], chain)
				nonce := util.StringToInt(util.U256(log.Topics[2]).String())
				member := GetMemberByAddress(ctx, wallet, chain)
				if member == nil {
					return errors.New("record not find")
				}

				db := util.DbBegin(ctx)

				amount := util.BigToDecimal(util.U256(log.Data))
				account := member.TouchAccount(ctx, currencyKton, wallet, chain)
				if err := account.subBalance(db, amount, "withdraw"); err != nil {
					db.DbRollback()
					return errors.Wrap(err, "balance insufficient")
				}
				err = member.updateWithdrawNonce(ctx, db, nonce+1, chain, currencyKton)
				if err != nil {
					db.DbRollback()
					return errors.Wrap(err, "new withdraw error")
				}
				th := TransactionHistory{Tx: ec.Tx, Chain: chain, BalanceAddress: wallet, BalanceChange: amount, Action: TransactionHistoryWithdraw, Currency: currencyKton}
				_ = th.New(db)
				if err = account.NewWithdraw(db, util.IntToString(nonce), ec.Tx, amount, currencyKton); err != nil {
					db.DbRollback()
					return errors.Wrap(err, "new withdraw error")
				}
				db.DbCommit()
				if db.Error != nil {
					return errors.Wrap(err, "commit TakeBackKtonCallback error")
				}
				go afterWithdraw(chain)
			}
		}
	}

	unTrDb := util.DbBegin(ctx)
	if err := ec.NewUniqueTransaction(unTrDb, "TakeBackKton"); err != nil {
		unTrDb.DbRollback()
		return err
	}
	unTrDb.DbCommit()
	return unTrDb.Error
}

func afterWithdraw(chain string) {
	if !util.IsProduction() {
		return
	}
	ring := newCurrency(currencyRing, chain)
	address := util.GetContractAddress("takeBack", chain)
	if balance := util.BigToDecimal(ring.GetBalance(address), util.GetTokenDecimals(chain)); balance.LessThan(decimal.New(2000000, 0)) {
		_ = services.NewIssue("NetWork: "+chain+" takeBack contract less than 2000000 ring, please deposit", "takeBack deposit")
	}
}
