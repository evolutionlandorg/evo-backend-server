package daemons

import (
	"context"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"strings"
	"time"

	"github.com/evolutionlandorg/block-scan/scan"
	"github.com/evolutionlandorg/block-scan/services"
	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func FreshBlockStatus(ctx context.Context) {
	defer util.Recover("FreshBlockStatus error")
	t := time.NewTicker(time.Second * 5)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug("FreshBlockStatus done")
			return
		case <-t.C:
			span, spanCtx := tracer.StartSpanFromContext(ctx, "daemons.worker",
				tracer.ServiceName("evo-backend-worker"),
				tracer.SpanType(ext.SpanTypeMessageConsumer),
				tracer.Measured(),
				tracer.Tag("worker-name", "FreshBlockStatus"),
			)
			for _, transaction := range models.GetEthTransactionPending(spanCtx) {
				chain := transaction.Chain
				if chain == "" {
					chain = "Eth"
					if !strings.HasPrefix(transaction.Tx, "0x") {
						chain = "Tron"
					}
				}
				_ = FreshTxStatus(spanCtx, transaction.Tx, chain)
				span.Finish()
				t.Reset(time.Second * 5)
			}
		}
	}
}

func FreshTxStatus(ctx context.Context, tx, chain string) error {
	sg := storage.New(chain)
	res, err := sg.ReceiptLog(tx)
	if err != nil || res == nil {
		return err
	}

	if sg.BlockNumber()-util.U256(res.BlockNumber).Uint64() <= 3 {
		return nil
	}

	status := sg.GetTransactionStatus(tx)
	et := models.EthTransaction{
		ReceiptsLog: util.ToString(res.Logs),
		Status:      status,
	}

	if status == "Success" {
		if len(res.Logs) == 0 {
			return nil
		}
		blockHeader := sg.BlockHeader(util.U256(res.BlockNumber).Uint64())
		if blockHeader == nil {
			return nil
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
	}
	et.UpdateEthTransaction(ctx, tx)
	return nil
}
