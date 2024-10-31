package commands

import (
	"context"
	"strings"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/urfave/cli"
)

var (
	SubAction = []cli.Command{
		{
			Name:  "refreshGeneValue",
			Usage: "refreshGeneValue",
			Action: func(c *cli.Context) error {
				RefreshGeneValue(context.TODO())
				return nil
			},
		},
		{
			Name:  "InitLandsFormChain",
			Usage: "InitLandsFormChain",
			Action: func(c *cli.Context) error {
				chain := c.Args().Get(0)
				models.InitLandsFormChain(context.TODO(), chain)
				models.AddLandsGameXY(context.TODO(), models.GetDistrictByChain(chain))
				return nil
			},
		},
		{
			Name:  "RefreshLands",
			Usage: "RefreshLands",
			Action: func(c *cli.Context) error {
				chain := c.Args().Get(0)
				models.RefreshLands(context.TODO(), chain)
				return nil
			},
		},
		{
			Name: "RefreshApostleTalent",
			Action: func(c *cli.Context) error {
				chain := c.Args().Get(0)
				tokenIds := c.Args().Get(1)
				refreshApostleTalent(context.TODO(), chain, tokenIds)
				return nil
			},
		},
		{
			Name: "RefreshDrillsFormulaId",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "chain",
					Required: true,
				},
				cli.StringFlag{
					Name:     "extClass",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				chain := c.String("chain")
				extClass := c.Int("extClass")
				return RefreshDrillsFormulaId(context.TODO(), chain, extClass)
			},
		},
		{
			Name: "RefreshTx",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "chain",
					Required: true,
				},
				cli.StringFlag{
					Name:     "tx",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				var txs = hashset.New()
				for _, v := range strings.Split(c.String("tx"), ",") {
					if v == "" {
						continue
					}
					if txs.Contains(v) {
						continue
					}
					txs.Add(v)
					if err := RefreshTxStatus(context.TODO(), c.String("chain"), v); err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			Name: "RebuildTransactionRecords",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name: "chain",
				},
				cli.BoolFlag{
					Name: "remove",
				},
			},
			Action: func(c *cli.Context) error {
				RebuildTransactionRecords(context.TODO(), RebuildTransactionRecordsOpt{
					Chain:      c.StringSlice("chain"),
					NeedRemove: c.Bool("remove"),
				})
				return nil
			},
		},
		{
			Name: "RefreshElementRaffle",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name: "chain",
				},
				cli.Int64SliceFlag{
					Name: "start_block",
				},
			},
			Action: func(c *cli.Context) error {
				return models.RefreshElementRaffle(context.TODO(), c.StringSlice("chain"), c.Int64Slice("start_block"))
			},
		},
	}
)
