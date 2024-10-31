package models

import (
	"context"
	"errors"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/shopspring/decimal"
)

func (ec *EthTransactionCallback) LootBoxCallback(ctx context.Context) (err error) {
	if getTransactionDeal(ctx, ec.Tx, "lootBox") != nil {
		return errors.New("tx exist")
	}

	txn := util.DbBegin(ctx)
	defer txn.DbRollback()
	chain := ec.Receipt.ChainSource

	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("lootBox", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			case services.AbiEncodingMethod("EVOHarbergerBuy(address,uint256,uint256,uint256)"):
				dataSlice := util.LogAnalysis(log.Data)
				address := util.AddHex(util.TrimHex(dataSlice[0])[24:64], chain)
				if err = NewTreasure(txn, address, ec.Tx, GoldBox, decimal.Zero, ec.BlockTimestamp, 1, GetChainById(int(util.U256(dataSlice[3]).Int64()))); err != nil {
					return err
				}
			}
		}
	}

	if err != nil {
		return err
	}

	if err := ec.NewUniqueTransaction(txn, "lootBox"); err != nil {
		return err
	}

	txn.DbCommit()
	return txn.Error
}
