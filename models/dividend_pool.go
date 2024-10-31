package models

import (
	"context"
	"errors"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
)

func (ec *EthTransactionCallback) DividendPoolCallback(ctx context.Context) (err error) {

	if getTransactionDeal(ctx, ec.Tx, "dividendPool") != nil {
		return errors.New("tx exist")
	}
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	exec := false
	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("dividendPool", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			case services.AbiEncodingMethod("TransferredChannelDividend(address,uint256)"):
				// logSlice := services.LogAnalysis(log.Data)
				// airDropFromKton(ec.Tx, util.U256(logSlice[0]))
				exec = true
			}
		}
	}
	if !exec {
		return nil
	}

	if err := ec.NewUniqueTransaction(db, "dividendPool"); err != nil {
		return err
	}

	db.DbCommit()
	return db.Error
}
