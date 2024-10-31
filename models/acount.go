package models

import (
	"context"
	"errors"
	"strings"

	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type Account struct {
	gorm.Model
	MemberId uint            `json:"member_id"`
	Currency string          `json:"currency"`
	Wallet   string          `json:"wallet"`
	Chain    string          `json:"chain"`
	Balance  decimal.Decimal `json:"balance" sql:"type:decimal(32,16); default:0"`
}

func (a *Account) AddBalance(db *util.GormDB, amount decimal.Decimal, reason string) error {
	if a != nil && amount.IsPositive() {
		if query := db.Model(a).Where("balance=?", a.Balance).Update(map[string]interface{}{"balance": gorm.Expr("balance + ?", amount)}); query == nil || query.RowsAffected == 0 {
			return errors.New("add balance fail")
		}
		if err := a.updateVersion(db, accountAdd, amount, reason); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("amount count error")
	}
}

func (a *Account) subBalance(db *util.GormDB, amount decimal.Decimal, reason string) error {
	if a != nil && amount.Sign() >= 0 && a.Balance.Cmp(amount) >= 0 {
		if query := db.Model(a).Where("balance=?", a.Balance).Update(map[string]interface{}{"balance": gorm.Expr("balance - ?", amount)}); query == nil || query.RowsAffected == 0 {
			return errors.New("sub balance fail")
		}
		a.Balance = a.Balance.Sub(amount)
		if err := a.updateVersion(db, accountSub, amount, reason); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("amount count error")
	}
}

func (a *Account) updateVersion(db *util.GormDB, op string, amount decimal.Decimal, reason string) error {
	av := AccountVersion{MemberId: a.MemberId, AccountId: a.ID, Currency: a.Currency}
	switch op {
	case accountAdd:
		av.Balance = amount
	case accountSub:
		av.Balance = amount.Neg()
	}
	sliceReason := strings.Split(reason, ":")
	if len(sliceReason) > 0 {
		av.Reason = sliceReason[0]
	}
	if len(sliceReason) > 1 {
		av.Remark = sliceReason[1]
	}
	return av.New(db)
}

func (a *Account) AccountRecordList(ctx context.Context, query VersionQuery) *[]AccountVersion {
	db := util.WithContextDb(ctx).Table("account_versions")
	var av []AccountVersion
	wheres, values := util.StructToSql(query)
	if len(wheres) != 0 {
		db = db.Where(strings.Join(wheres, " AND "), values...)
	}
	db.Find(&av)
	return &av
}

func addRewardOnce(ctx context.Context, db *util.GormDB, address, reason, remark string, reward decimal.Decimal, chain string, currency string) error {
	member := GetMemberByAddress(ctx, address, chain)
	if member == nil {
		member = &Member{Wallet: address}
	}
	if currency == "" {
		currency = currencyRing
	}
	account := member.TouchAccount(ctx, currency, address, chain)
	if len(*account.AccountRecordList(ctx, VersionQuery{Reason: reason, AccountId: account.ID, Remark: remark})) == 0 {
		return account.AddBalance(db, reward, reason+":"+remark)
	}
	return nil
}
