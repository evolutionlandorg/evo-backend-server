package models

import (
	"context"
	"errors"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
)

func (ec *EthTransactionCallback) DrillTakeBackCallback(ctx context.Context) (err error) {
	if getTransactionDeal(ctx, ec.Tx, "DrillTakeBack") != nil {
		return errors.New("tx exist")
	}

	txn := util.DbBegin(ctx)
	defer txn.DbRollback()
	chain := ec.Receipt.ChainSource
	var address string
	// var nonce int
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("DrillTakeBack", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			case services.AbiEncodingMethod("OpenBox(address,uint256,uint256,uint256)"):
				logSlice := util.LogAnalysis(log.Data)
				address = util.AddHex(util.TrimHex(log.Topics[1])[24:64], chain)
				boxId := util.Padding(util.BytesToHex(util.U256(log.Topics[2]).Bytes()))
				tokenId := logSlice[0]
				value := util.BigToDecimal(util.U256(logSlice[1]), util.GetTokenDecimals(chain))
				err = openGen2Treasure(txn, boxId, tokenId, value, ec.BlockTimestamp)

				// history record
				th := TransactionHistory{Tx: ec.Tx, Chain: chain, BalanceAddress: address, Action: TransactionHistoryLuckyBoxOpen, BalanceChange: value, Currency: currencyRing, TokenId: tokenId}
				_ = th.New(txn)
				// TakeBackDrill(address indexed user, uint256 indexed id, uint256 tokenId)
			case services.AbiEncodingMethod("TakeBackDrill(address,uint256,uint256)"):
				address = util.AddHex(util.TrimHex(log.Topics[1])[24:64], chain)
				logSlice := util.LogAnalysis(log.Data)
				memberId := util.U256(log.Topics[2]).Int64()
				if member := GetMember(ctx, int(memberId)); member != nil {
					member.Newbie = "rewarded"
					_ = member.updateField(ctx, txn)
				}
				// history record
				th := TransactionHistory{Tx: ec.Tx, Chain: chain, BalanceAddress: address, Action: TransactionHistoryNewbieReward, TokenId: logSlice[0]}
				_ = th.New(txn)
			}
		}
	}

	if err := ec.NewUniqueTransaction(txn, "DrillTakeBack"); err != nil {
		return err
	}

	txn.DbCommit()
	return txn.Error
}
