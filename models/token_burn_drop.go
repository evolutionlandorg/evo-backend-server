package models

import (
	"context"
	"errors"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
)

func (ec *EthTransactionCallback) TokenBurnDropCallback(ctx context.Context) (err error) {
	if getTransactionDeal(ctx, ec.Tx, "TokenBurnDrop") != nil {
		return errors.New("tx exist")
	}

	db := util.DbBegin(ctx)
	defer db.DbRollback()

	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("tokenBurnDrop", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			case services.AbiEncodingMethod("RingBurndropTokens(address,address,uint256,bytes)"):
				fallthrough
			case services.AbiEncodingMethod("KtonBurndropTokens(address,address,uint256,bytes)"):
				// send SendReceiptsProofToDarwinia
				logSlice := util.LogAnalysis(log.Data)
				address := util.AddHex(util.TrimHex(log.Topics[2])[24:64], chain)
				token := util.AddHex(util.TrimHex(log.Topics[1])[24:64], chain)
				mount := util.BigToDecimal(util.U256(logSlice[0]))

				th := TransactionHistory{Tx: ec.Tx, Chain: chain, BalanceAddress: address, Action: TransactionHistoryKtonMapping, BalanceChange: mount, Currency: currencyKton}
				if token == util.GetContractAddress("ring", chain) {
					th.Action = TransactionHistoryRingMapping
					th.Currency = currencyRing
				}
				// tokenMapToDarwinia(ec.Tx, th.Currency)
				_ = th.New(db)
			}
		}
	}

	if err != nil {
		return err
	}

	if err := ec.NewUniqueTransaction(db, "TokenBurnDrop"); err != nil {
		return err
	}

	db.DbCommit()
	return db.Error
}
