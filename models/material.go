package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/evolutionlandorg/evo-backend/util/pve"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/shopspring/decimal"
)

func (ec *EthTransactionCallback) MaterialTakeBackCallback(ctx context.Context) error {
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	if getTransactionDeal(ctx, ec.Tx, "materialTakeBack") != nil {
		return errors.New("tx exist")
	}
	chain := ec.Receipt.ChainSource
	var (
		nonce  int
		member *Member
	)

	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) == 0 || !strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("materialTakeBack", chain)) {
			continue
		}
		eventName := util.AddHex(log.Topics[0])
		switch eventName {
		// [account,nonce,id,tokenId,amount]
		case services.AbiEncodingMethod("TakebackMaterial(address,uint256,uint128,uint256,uint256)"):
			logSlice := util.LogAnalysis(log.Data)
			wallet := util.AddHex(logSlice[0][24:64], chain)
			nonce = util.StringToInt(util.U256(logSlice[1]).String())
			member = GetMemberByAddress(ctx, wallet, chain)
			if member == nil {
				return errors.New("record not find")
			}
			amount := decimal.NewFromBigInt(util.U256(logSlice[4]), 0)
			materialId := util.StringToInt(util.U256(logSlice[2]).String())
			if err := materialTakeBack(db, ec.Tx, wallet, chain, materialId, amount); err != nil {
				return err
			}
		}
	}
	if member != nil {
		if err := member.UpdateMaterialNonce(ctx, chain, nonce); err != nil {
			return err
		}
	}

	if err := ec.NewUniqueTransaction(db, "materialTakeBack"); err != nil {
		return err
	}

	db.DbCommit()
	return db.Error
}

func materialTakeBack(db *util.GormDB, tx, wallet, chain string, materialId int, _ decimal.Decimal) error {
	currency := pve.MaterialIdToSymbol(materialId)
	if currency == "" {
		return fmt.Errorf("unknown materialId %d", materialId)
	}
	member := GetMemberByAddress(db.Context(), wallet, chain)
	if member == nil {
		return errors.New("record not find")
	}

	for _, v := range []string{CrabChain, EthChain, HecoChain, PolygonChain, TronChain} { // pvp 新版本无视chain, 只要领取了全部清空
		account := member.TouchAccount(db.Context(), currency, wallet, v, true)
		if account == nil {
			continue
		}
		if account.Balance.LessThanOrEqual(decimal.NewFromInt(0)) {
			continue
		}
		if err := account.subBalance(db, account.Balance, ReasonWithdrawMaterial); err != nil {
			return errors.New("balance insufficient")
		}
		th := TransactionHistory{Tx: tx, Chain: chain, BalanceAddress: wallet, BalanceChange: account.Balance, Action: TransactionHistoryWithdraw, Currency: currency}
		if err := th.New(db); err != nil {
			return err
		}
	}
	return nil
}
