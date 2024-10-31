package models

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/orcaman/concurrent-map"
	"github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

func (ec *EthTransactionCallback) PointsRewardCallback(ctx context.Context) (err error) {
	if getTransactionDeal(ctx, ec.Tx, "pointsReward") != nil {
		return errors.New("tx exist")
	}
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	var (
		address    string
		jackpotWin int64
	)
	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("pointsReward", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			case services.AbiEncodingMethod("RewardClaimedWithPoints(address,uint256,uint256)"):
				address = util.AddHex(util.TrimHex(log.Topics[1])[24:64], chain)
				logSlice := util.LogAnalysis(log.Data)
				pointAmount := util.BigToDecimal(util.U256(logSlice[0]), util.GetTokenDecimals(chain))
				rewardAmount := util.BigToDecimal(util.U256(logSlice[1]), util.GetTokenDecimals(chain))
				jackpotWin = util.BigToDecimal(util.U256(logSlice[1]), util.GetTokenDecimals(chain)).IntPart()
				th := TransactionHistory{Tx: ec.Tx, Chain: chain, BalanceAddress: address, Action: TransactionHistoryTickets,
					Currency: currencyRing, BalanceChange: rewardAmount.Round(8),
					Extra: "smallTickets",
				}
				if pointAmount.String() == "100" {
					th.Extra = "largeTickets"
				}
				_ = th.New(db)
			}
		}
	}

	if err != nil {
		return err
	}

	if err := ec.NewUniqueTransaction(db, "pointsReward"); err != nil {
		return err
	}

	db.DbCommit()
	if db.Error != nil {
		return db.Error
	}

	if jackpotWin >= 10000 {
		pointRewardBroadcast(ctx, address, jackpotWin, chain)
	}
	return err
}

func PointPollMap(address, chain string) map[string]interface{} {
	resultsMap := cmap.New()
	var wg sync.WaitGroup
	origin := []string{"poll", "mine"}
	sg := storage.New(chain)
	for _, op := range origin {
		wg.Add(1)
		go func(op string) {
			defer wg.Done()
			var balance decimal.Decimal
			if op == "poll" {
				balance = sg.TotalRewardInPool()
			} else {
				balance = sg.PointsBalanceOf(address)
			}
			resultsMap.Set(op, balance)
		}(op)
	}
	wg.Wait()
	return resultsMap.Items()
}

func pointRewardBroadcast(ctx context.Context, winner string, jackpotWin int64, chain string) {
	name := "alien"
	if member := GetMemberByAddress(ctx, winner, chain); member != nil {
		name = member.Name
	}
	content := fmt.Sprintf("%s|%d", name, jackpotWin)
	broadcastChannel(ctx, content, BroadcastPointReward)
	bm := BroadcastMessage{Content: content, MsgId: uuid.Must(uuid.NewV4(), nil).String(), ExpiredAt: int(time.Now().Unix()) + 120, ContentType: BroadcastPointReward}
	_ = bm.New(ctx)
}
