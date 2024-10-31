// Package daemons provides ...
package daemons

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/evolutionlandorg/staker/apr"
	"github.com/shopspring/decimal"
)

type FarmStorage struct {
	//pool string
	//lp   string
	//s    storage.IStorage
}

var (
	DECIMAIL = 2
	Pools    = []string{
		"lpGoldPool",
		"lpWoodPool",
		"lpWaterPool",
		"lpFirePool",
		"lpSoilPool",
		"ethPool",
		"lpKtonPool",
		"lpDUSDPool",
	}

	Big1 = big.NewInt(1)
)

func FreshFarmAPR(ctx context.Context) {
	t := time.NewTicker(time.Minute)
	for {
		select {
		case <-ctx.Done():
			log.Debug("FreshFarmAPR done")
			return
		case <-t.C:
			FreshChainFarmAPR(ctx, "Heco")
			FreshChainFarmAPR(ctx, "Polygon")
			FreshChainFarmAPR(ctx, "Crab")
			t.Reset(time.Minute)
		}
	}
}

func FreshChainFarmAPR(ctx context.Context, chain string) {
	defer util.Recover(fmt.Sprintf("FreshFarmAPR %s error", chain), true)
	var (
		s                 = storage.New(chain)
		c                 = apr.New(s, DECIMAIL)
		ring              = util.GetContractAddress("ring", chain)
		kton              = util.GetContractAddress("kton", chain)
		removeInvalidTime = time.Now().Add(-time.Hour * 24 * 7)
	)

	for _, pool := range Pools {
		p := util.GetContractAddress(pool, chain)
		if p == "" {
			continue
		}
		base := ""
		transformer := apr.NewFraction(Big1, Big1)
		if pool == "lpKtonPool" {
			base = kton
			ringPrice, err := services.GetRingPrice()
			if err != nil {
				log.Error("FreshChainFarmAPR GetRingPrice failed. chain %s, pool %s, error: %s",
					chain, pool, err)
				continue
			}
			ktonPrice, err := services.GetKtonPrice()
			if err != nil {
				log.Error("FreshChainFarmAPR GetKtonPrice failed. chain %s, pool %s, error: %s",
					chain, pool, err)
				continue
			}
			transformer = apr.NewFraction(
				ktonPrice.Mul(decimal.NewFromInt(100000000)).BigInt(),
				ringPrice.Mul(decimal.NewFromInt(100000000)).BigInt(),
			)
		} else {
			base = ring
		}
		a, err := c.Calc(p, base, ring, transformer)
		if err != nil {
			log.Error("FreshChainFarmAPR failed. chain %s, pool %s, error: %s",
				chain, pool, err)
			continue
		}
		if err := models.RawAddFarmAPR(ctx, pool, p, fmt.Sprintf("%.2f", float32(a))); err != nil {
			log.Error("DB Insert APR failed. chain %s, pool %s, error: %s",
				chain, pool, err)
		}
		if err := models.RemoveFarmAPRByTime(ctx, p, removeInvalidTime); err != nil {
			log.Error("remove invalid APR data failed. chain %s, pool %s, error: %s",
				chain, pool, err)
		}
	}
}
