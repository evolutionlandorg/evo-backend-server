package daemons

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	block_scan "github.com/evolutionlandorg/block-scan"
	"github.com/evolutionlandorg/block-scan/services"
	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/gomodule/redigo/redis"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cast"
)

func Start(ctx context.Context) {
	fns := []func(ctx context.Context){
		FreshBlockStatus,
		FreshFarmAPR,
		FreshSwapStatus,
		StartUploadData,
		StartWorker,
		StartWipeBlock,
		StartSnapshot,
	}

	do := func(fn func(ctx context.Context)) func() {
		return func() {
			fn(ctx)
		}
	}
	for _, fn := range fns {
		go util.RecoverRunForever("freshBlockStatus error", do(fn), time.Second*10, true)
	}
}

func StartWipeBlock(ctx context.Context) {
	do := func(chain string, contractsMap util.ContractAddress) func() {
		return func() {
			startWipeTrxBlock(ctx, chain, contractsMap)
		}
	}
	for chain, contractsMap := range util.Evo.Contracts {
		if util.IsProduction() && chain == storage.Bsc {
			continue
		}

		if !util.IsProduction() && util.IntInSlice(chain, []string{storage.Ethereum}) {
			continue
		}
		go util.RecoverRunForever(
			fmt.Sprintf("%s WipeBlock error", chain),
			do(chain, contractsMap),
			time.Second*10,
			true,
		)
	}
}

func StartWorker(ctx context.Context) {
	go util.RecoverRunForever("worker error", RunWorker, time.Second*10, true)
}

func startWipeTrxBlock(ctx context.Context, chain string, contractsMap util.ContractAddress) {
	var scanType = block_scan.SUBSCRIBE
	if chain == models.TronChain {
		scanType = block_scan.POLLING
	}
	var contractsName = make(map[services.ContractsAddress]services.ContractsName)
	for address, name := range contractsMap {
		contractsName[services.ContractsAddress(strings.ToLower(address))] = services.ContractsName(name)
	}
	_ = os.Setenv(fmt.Sprintf("%s_WSS_RPC", strings.ToUpper(chain)), storage.GetChainWssRpc(chain))
	util.Panic(block_scan.StartScanChainEvents(ctx, scanType, services.ScanEventsOptions{
		ChainIo: storage.New(chain),
		GetStartBlock: func() uint64 {
			n, _ := redis.Uint64(util.SubPoolWithContextDo(context.TODO())("HGET", "WipeBlock", chain))
			return n
		},
		SetStartBlock: func(currentBlockNum uint64) {
			_, _ = util.SubPoolWithContextDo(context.TODO())("HSET", "WipeBlock", chain, currentBlockNum)
		},
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
				return true
			},
		},
	}))
}
