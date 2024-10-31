package models

import (
	"context"
	"errors"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
)

func (ec *EthTransactionCallback) ItemBaseCallback(ctx context.Context) (err error) {
	if getTransactionDeal(ctx, ec.Tx, "ItemBase") != nil {
		return errors.New("tx exist")
	}

	txn := util.DbBegin(ctx)
	defer txn.DbRollback()
	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("ItemBase", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {

			// indexed user, indexed tokenId, index, rate, objClassExt, class, grade, prefer, major, id, minor, amount, now
			case services.AbiEncodingMethod("Enchanced(address,uint256,uint256,uint128,uint16,uint16,uint16,uint16,address,uint256,address,uint256,uint256)"):
				logSlice := util.LogAnalysis(log.Data)
				address := util.AddHex(util.TrimHex(log.Topics[1])[24:64], chain)
				tokenId := util.TrimHex(log.Topics[2])
				formulaIndex := util.StringToInt(util.U256(logSlice[0]).String())

				objClassExt := util.StringToInt(util.U256(logSlice[2]).String())
				class := util.StringToInt(util.U256(logSlice[3]).String())
				grade := util.StringToInt(util.U256(logSlice[4]).String())
				prefer := util.StringToInt(util.U256(logSlice[5]).String())
				createTime := util.StringToInt(util.U256(logSlice[10]).String())

				err := CreateDrill(txn, address, tokenId, createTime, class, grade, prefer, formulaIndex, objClassExt, chain)

				if err != nil {
					return err
				}
				// history record
				th := TransactionHistory{Tx: ec.Tx, Chain: chain, BalanceAddress: address, Action: DrillEnchanced, TokenId: tokenId}
				_ = th.New(txn)
			// indexed user,tokenId,address,major, id, minor, amount
			case services.AbiEncodingMethod("Disenchanted(address,uint256,address,uint256,address,uint256)"):

				logSlice := util.LogAnalysis(log.Data)
				address := util.AddHex(util.TrimHex(log.Topics[1])[24:64], chain)
				tokenId := util.TrimHex(logSlice[0])
				// history record
				th := TransactionHistory{Tx: ec.Tx, Chain: chain, BalanceAddress: address, Action: DrillDisenchanted, TokenId: tokenId}
				_ = th.New(txn)
			}
		}
	}

	if err != nil {
		return err
	}

	if err := ec.NewUniqueTransaction(txn, "ItemBase"); err != nil {
		return err
	}

	txn.DbCommit()
	return txn.Error
}
