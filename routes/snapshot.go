package routes

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

// Snapshot docs
// @Summary	list snapshot vote score
// @Produce	json
// @Tags		common
// @Param		options	body		models.SnapshotReq							true	"options"
// @Success	200		{object}	routes.GinJSON{data=models.SnapshotAsJson}	"success"
// @Router		/snapshot [post]
func Snapshot() func(c *gin.Context) {
	return func(c *gin.Context) {
		var args models.SnapshotReq
		if err := c.Bind(&args); err != nil {
			getReturnDataByError(c, 10001, err.Error())
			return
		}
		if args.Options.Land <= 0 {
			args.Options.Land = 100
		}
		if args.Options.Apostle <= 0 {
			args.Options.Apostle = 1
		}

		if len(args.Options.Chain) <= 0 {
			args.Options.Chain = append(args.Options.Chain, models.EthChain, models.CrabChain, models.HecoChain, models.PolygonChain, models.TronChain)
		}

		args.Network = util.GetNetworkNameById(args.Network)
		if args.Network == "" {
			getReturnDataByError(c, 10001)
			return
		}
		var oldWallet = make(map[string]string, len(args.Addresses))
		for _, v := range args.Addresses {
			oldWallet[strings.ToLower(v)] = v
		}
		var (
			wg        sync.WaitGroup
			otherVote sync.Map
		)
		for _, chain := range args.Options.Chain {
			for _, v := range args.Addresses {
				if args.Options.Kton > 0 {
					wg.Add(1)
					go func(address, chain string) {
						defer wg.Done()
						sg := storage.New(chain)
						_ = util.Try(func() error {
							balance := util.BigToDecimal(sg.BalanceOf(address, util.GetContractAddress("kton", models.HecoChain)))
							if balance.IsNegative() {
								return errors.New("rpc error")
							}

							otherVote.Store(fmt.Sprintf("%s-kton", address), balance.IntPart())
							return nil
						}, 10)
					}(v, chain)
				}
				if args.Options.Element > 0 {
					wg.Add(1)
					go func(address, chain string) {
						defer wg.Done()
						sg := storage.New(chain)
						var balance decimal.Decimal
						for _, v := range []string{"gold", "wood", "water", "fire", "soil"} {
							_ = util.Try(func() error {
								elementBalance := util.BigToDecimal(sg.BalanceOf(address, util.GetContractAddress(v, models.HecoChain)))
								if elementBalance.IsNegative() {
									return errors.New("rpc error")
								}
								balance = balance.Add(elementBalance)
								return nil
							}, 10)
						}
						otherVote.Store(fmt.Sprintf("%s-element", address), balance.IntPart())
					}(v, chain)
				}
			}
		}

		result, err := models.GetSnapshotByAddress(util.GetContextByGin(c), args.Addresses, cast.ToInt(args.Snapshot), args.Options.Chain, args.Network)
		if err != nil {
			log.Error("get snapshot error %s", err)
			c.JSON(200, []string{})
			return
		}
		wg.Wait()
		var (
			ktonVotes    = make(map[string]float64)
			elementVotes = make(map[string]float64)
		)

		otherVote.Range(func(key, value interface{}) bool {
			address := strings.Split(key.(string), "-")
			if address[1] == "element" {
				elementVotes[address[0]] += args.Options.Element * float64(value.(int64))
				return true
			}
			ktonVotes[address[0]] += args.Options.Kton * float64(value.(int64))
			return true
		})
		var snapshots = new(models.SnapshotAsJson)
		for k, wallet := range oldWallet {
			sr := new(models.SnapshotResp)
			if v, ok := result[k]; ok {
				sr.Land = v.LandNumber
				sr.Apostle = v.ApostleNumber
			}
			snapshots.Score = append(snapshots.Score, models.Score{
				Score:   args.Options.Apostle*sr.Apostle + args.Options.Land*sr.Land + int64(elementVotes[wallet]) + int64(ktonVotes[wallet]),
				Address: wallet,
			})
		}
		c.JSON(200, snapshots)
	}
}
