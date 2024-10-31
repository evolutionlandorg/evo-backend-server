package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	block_scan "github.com/evolutionlandorg/block-scan"
	"github.com/evolutionlandorg/block-scan/scan"
	"github.com/evolutionlandorg/block-scan/services"
	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cast"
)

func RefreshGeneValue(ctx context.Context) {
	db := util.WithContextDb(ctx)
	var list []models.Apostle
	db.Model(models.Apostle{}).Find(&list)
	for _, apostle := range list {
		rb := util.HttpGet(fmt.Sprintf("%s/gn%s.svge", util.ApiServerHost, util.U256(apostle.Genes).String()))
		var svgAttributes []models.SvgServerAttribute
		err := json.Unmarshal(rb, &svgAttributes)
		if err != nil || len(svgAttributes) < 1 {
			log.Debug("get data by %s error: %s", string(rb), err)
			return
		}
		for _, v := range svgAttributes {
			kind := models.SvgIdToKindMap[v.Id]
			if kind == "" {
				continue
			}
			var attribute models.Attribute
			query := db.First(&attribute, map[string]interface{}{"title": v.Name, "kind": kind})
			if !query.RecordNotFound() {
				if attribute.Value == 0 {
					db.Model(&attribute).UpdateColumn(map[string]interface{}{"value": v.Value})
				}
			}
		}
	}
}

func refreshApostleTalent(ctx context.Context, chain string, tokenIds string) {
	sg := storage.New(chain)
	tokenIdArr := strings.Split(tokenIds, ",")
	for _, tokenId := range tokenIdArr {
		if apostleProp := sg.TokenId2Apostle(tokenId); len(apostleProp) > 2 {
			txn := util.DbBegin(ctx)
			apostle := models.Apostle{TokenId: tokenId}
			_ = apostle.RefreshTalent(txn, apostleProp[1])
			txn.DbCommit()
		}
	}
}

func RefreshDrillsFormulaId(ctx context.Context, chain string, extClass int) error {
	var drills []models.Drill
	db := util.WithContextDb(ctx)
	db = db.Model(models.Drill{}).Where("chain = ? AND formula_id = 0", chain).Find(&drills)
	if db.Error != nil {
		return db.Error
	}
	for _, drill := range drills {
		if err := drill.RefreshFormulaId(ctx, extClass); err != nil {
			return err
		}
	}
	return nil
}

func RefreshTxStatus(ctx context.Context, chain, tx string) error {
	sg := storage.New(chain)
	var res *services.Receipts
	if err := util.Try(func() error {
		var err error
		res, err = sg.ReceiptLog(tx)
		return err
	}, 3); err != nil {
		return fmt.Errorf("ReceiptLog error %s", err)
	}
	if sg.BlockNumber()-util.U256(res.BlockNumber).Uint64() <= 3 {
		return nil
	}

	et := models.EthTransaction{
		ReceiptsLog: util.ToString(res.Logs),
		Status:      res.Status,
	}
	if res.Status != "0x1" {
		return errors.New("tx status is not Success")
	}
	if len(res.Logs) == 0 {
		return errors.New("tx log is empty")
	}
	var blockHeader *services.BlockHeader
	if err := util.Try(func() error {
		blockHeader = sg.BlockHeader(util.U256(res.BlockNumber).Uint64())
		if blockHeader == nil {
			return errors.New("rpc error")
		}
		return nil
	}, 3); err != nil {
		return err
	}

	var contractsName = make(map[services.ContractsAddress]services.ContractsName)
	for address, name := range util.GetContractsMap(chain) {
		contractsName[services.ContractsAddress(strings.ToLower(address))] = services.ContractsName(name)
	}
	p := &scan.Polling{
		Opt: services.ScanEventsOptions{
			CallbackMethodPrefix: util.Evo.ContractsListen,
			ContractsName:        contractsName,
			Chain:                chain,
			GetCallbackFunc: func(tx string, blockTimestamp uint64, receipt *services.Receipts) interface{} {
				return models.EthTransactionCallback{
					Tx:             tx,
					Receipt:        receipt,
					BlockTimestamp: int64(blockTimestamp),
				}
			},
		},
	}
	_ = p.ReceiptDistribution(tx, blockHeader.BlockTimeStamp, res)

	et.UpdateEthTransaction(ctx, tx)
	return nil
}

type RebuildTransactionRecordsOpt struct {
	Chain      []string
	NeedRemove bool
}

func RebuildTransactionRecords(ctx context.Context, opt RebuildTransactionRecordsOpt) {
	if len(opt.Chain) == 0 {
		opt.Chain = append(opt.Chain, models.CrabChain, models.EthChain, models.TronChain, models.HecoChain, models.PolygonChain)
	}

	for _, v := range opt.Chain {
		var startBlock uint64
		db := util.WithContextDb(ctx)
		if opt.NeedRemove {
			util.Panic(db.Where("chain = ?", v).Delete(new(models.TransactionScan)).Error)
			startBlock = util.Evo.WipeBlock[v].InitBlock
		}

		if !opt.NeedRemove {
			var tr models.TransactionScan
			db.Where("chain = ?", v).Order("block_number DESC").Limit(1).Find(&tr)
			startBlock = uint64(tr.BlockNumber)
		}

		if startBlock <= 0 {
			startBlock = util.Evo.WipeBlock[v].InitBlock
		}

		go func(chain string, startBlock uint64) {
			var scanType = block_scan.SUBSCRIBE
			if chain == models.TronChain {
				scanType = block_scan.POLLING
			}
			var contractsName = make(map[services.ContractsAddress]services.ContractsName)
			for address, name := range util.Evo.Contracts[chain] {
				contractsName[services.ContractsAddress(strings.ToLower(address))] = services.ContractsName(name)
			}
			_ = os.Setenv(fmt.Sprintf("%s_WSS_RPC", strings.ToUpper(chain)), storage.GetChainWssRpc(chain))
			util.Panic(block_scan.StartScanChainEvents(context.TODO(), scanType, services.ScanEventsOptions{
				ChainIo: storage.New(chain),
				GetStartBlock: func() uint64 {
					return startBlock
				},
				SetStartBlock: func(currentBlockNum uint64) {},
				Chain:         chain,
				ContractsName: contractsName,
				SleepTime:     5,
				GetCallbackFunc: func(tx string, blockTimestamp uint64, receipt *services.Receipts) interface{} {
					return models.EthTransactionCallback{
						Tx:             tx,
						Receipt:        receipt,
						BlockTimestamp: int64(blockTimestamp),
					}
				},
				CallbackMethodPrefix: util.Evo.ContractsListen,
				InitBlock:            util.Evo.WipeBlock[chain].InitBlock,
				RunForever:           true,
				BeforePushMiddleware: []services.BeforePushFunc{
					func(tx string, BlockTimestamp uint64, receipt *services.Receipts) bool {
						ts := &models.TransactionScan{
							Tx:             tx,
							Chain:          chain,
							BlockNumber:    cast.ToInt64(receipt.BlockNumber),
							BlockTimestamp: int64(BlockTimestamp),
						}
						data, _ := jsoniter.Marshal(receipt.Logs)
						ts.Logs = string(data)
						ts.New(context.TODO())
						return false
					},
				},
			}))
		}(v, startBlock)
	}
	select {}
}
