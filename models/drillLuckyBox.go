package models

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/shopspring/decimal"
)

func (ec *EthTransactionCallback) DrillLuckyBoxCallback(ctx context.Context) (err error) {

	if getTransactionDeal(ctx, ec.Tx, "DrillLuckyBox") != nil {
		return errors.New("tx exist")
	}

	txn := util.DbBegin(ctx)
	defer txn.DbRollback()
	chain := ec.Receipt.ChainSource

	var deal = func(topics []string, data, boxType string) error {
		logSlice := util.LogAnalysis(data)
		address := util.AddHex(util.TrimHex(topics[1])[24:64], chain)
		amount := util.U256(logSlice[0]).Int64()
		price := util.BigToDecimal(util.U256(logSlice[1]), util.GetTokenDecimals(chain))
		err = NewTreasure(txn, address, ec.Tx, boxType, price, ec.BlockTimestamp, amount, chain)

		// history record
		th := TransactionHistory{Tx: ec.Tx, Chain: chain, BalanceAddress: address, Action: TransactionHistoryLuckyBox, BalanceChange: price.Mul(decimal.New(amount, 0)).Neg(), Currency: currencyRing, Extra: fmt.Sprintf(`{"LuckyBox":"%s"}`, boxType)}
		_ = th.New(txn)

		return err
	}

	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("DrillLuckyBox", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			case services.AbiEncodingMethod("GoldBoxSale(address,uint256,uint256)"):
				err = deal(log.Topics, log.Data, "gold")
			case services.AbiEncodingMethod("SilverBoxSale(address,uint256,uint256)"):
				err = deal(log.Topics, log.Data, "silver")
			}
			if err != nil {
				return err
			}
		}
	}

	if err := ec.NewUniqueTransaction(txn, "DrillLuckyBox"); err != nil {
		return err
	}

	txn.DbCommit()
	return txn.Error
}
