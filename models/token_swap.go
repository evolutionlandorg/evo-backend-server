package models

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
	log1 "github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type TokenSwap struct {
	gorm.Model
	From          string          `json:"from"`
	To            string          `json:"to"`
	SwapContract  string          `json:"swap_contract"`
	ChainPair     string          `json:"chain_pair"`
	SwapTx        string          `json:"swap_tx"`
	FinishTx      string          `json:"finish_tx"`
	Status        string          `json:"status"`
	Currency      string          `json:"currency"`
	Fee           decimal.Decimal `json:"fee" sql:"type:decimal(32,18);"`
	Amount        decimal.Decimal `json:"amount" sql:"type:decimal(32,18);"`
	Confirmations int             `json:"confirmations"`
}

const (
	tokenSwapFinish    = "Finish"
	tokenSwapConfirmed = "confirmed"
)

var swapChainPair = map[string]string{"Eth": "EthTron", "Tron": "TronEth"}

func (ts *TokenSwap) new(txn *util.GormDB) error {
	ts.Status = tokenSwapConfirmed
	ts.Confirmations = 0
	query := txn.Create(&ts)
	return query.Error
}

func (ts *TokenSwap) UpdateSwapTx(ctx context.Context, confirmation int, chain string) error {
	txn := util.DbBegin(ctx)
	defer txn.DbRollback()
	extra := map[string]interface{}{"to": ts.To, "confirmations": confirmation, "max_confirmations": 10}
	bExtra, _ := json.Marshal(extra)
	setTransactionHistoryExtra(txn, ts.SwapTx, TransactionHistorySwap, string(bExtra))
	if confirmation >= 10 {
		updateField := TokenSwap{Confirmations: confirmation, Status: tokenSwapFinish}
		if query := txn.Model(&ts).UpdateColumn(updateField); query.Error != nil || query.RowsAffected < 1 {
			txn.Rollback()
			return errors.New("update fail")
		}
		ts.finishSwap(ctx, txn, chain)
	}
	txn.DbCommit()
	return nil
}

func (ts *TokenSwap) finishSwap(ctx context.Context, txn *util.GormDB, chain string) {
	_ = addRewardOnce(ctx, txn, ts.To, reasonSwap, ts.SwapTx, ts.Amount, chain, ts.Currency)
}

func (ec *EthTransactionCallback) TokenSwapCallback(ctx context.Context) (err error) {
	if getTransactionDeal(ctx, ec.Tx, "tokenSwap") != nil {
		return errors.New("tx exist")
	}
	txn := util.DbBegin(ctx)
	defer txn.DbRollback()
	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("tokenSwap", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			case services.AbiEncodingMethod("TokenSwapped(uint256,address,bytes32,uint256,address,uint256,uint256,uint256)"):
				chainPair := swapChainPair[chain]
				sliceData := util.LogAnalysis(log.Data)
				from := util.AddHex(sliceData[0][24:64], chain)
				var to string
				if chain == "Eth" {
					to = util.AddHex(sliceData[1][24:64], "Tron")
				} else {
					to = util.AddHex(sliceData[1][24:64], "Eth")
				}
				token := util.AddHex(sliceData[3][24:64], chain)

				var currency string
				if strings.EqualFold(token, util.GetContractAddress(currencyKton, chain)) {
					currency = currencyKton
				} else if strings.EqualFold(token, util.GetContractAddress(currencyRing, chain)) {
					currency = currencyRing
				} else {
					log1.Debug("not support currency swap. chain=%s. token=%s. currencyKton=%s. currencyRing=%s",
						chain, token, currencyKton, currencyRing)
					return
				}
				amount := util.BigToDecimal(util.U256(sliceData[2]))
				fee := util.BigToDecimal(util.U256(sliceData[4]))
				ts := TokenSwap{From: from, To: to, Amount: amount, Fee: fee, Confirmations: 0,
					SwapContract: util.GetContractAddress("tokenSwap", chain), SwapTx: ec.Tx, Status: tokenSwapConfirmed, ChainPair: chainPair, Currency: currency}
				_ = ts.new(txn)
				extra := map[string]interface{}{"to": to, "confirmations": 0, "max_confirmations": 10}
				bExtra, _ := json.Marshal(extra)
				th := TransactionHistory{Tx: ec.Tx, Chain: chain, BalanceAddress: from, Action: TransactionHistorySwap, BalanceChange: amount.Neg(), Currency: currency, Extra: string(bExtra)}
				_ = th.New(txn)
			}
		}
	}

	if err != nil {
		return err
	}

	if err := ec.NewUniqueTransaction(txn, "tokenSwap"); err != nil {
		return err
	}

	txn.DbCommit()
	return txn.Error
}

func NeedToFreshSwapTx(ctx context.Context) []TokenSwap {
	db := util.WithContextDb(ctx)
	var ut []TokenSwap
	query := db.Where("status = ?", tokenSwapConfirmed).Find(&ut)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	return ut
}
