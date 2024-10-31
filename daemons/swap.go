package daemons

import (
	"context"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"time"

	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func FreshSwapStatus(ctx context.Context) {
	defer util.Recover("FreshSwapStatus error")
	t := time.NewTicker(time.Second * 5)
	defer t.Stop()

	var chain string
	storageMap := map[string]storage.IStorage{
		"EthTron": storage.New("Eth"),
		"TronEth": storage.New("Tron"),
	}

	for {

		select {
		case <-ctx.Done():
			log.Debug("FreshSwapStatus done")
			return
		case <-t.C:
			span, spanCtx := tracer.StartSpanFromContext(ctx, "daemons.worker",
				tracer.ServiceName("evo-backend-worker"),
				tracer.SpanType(ext.SpanTypeMessageConsumer),
				tracer.Measured(),
				tracer.Tag("worker-name", "FreshSwapStatus"),
				tracer.Tag("chain", chain),
			)
			list := models.NeedToFreshSwapTx(spanCtx)
			for _, tx := range list {
				switch tx.ChainPair {
				case "EthTron":
					chain = "Tron"
				case "TronEth":
					chain = "Eth"
				default:
					span.Finish()
					continue
				}

				sg := storageMap[tx.ChainPair]
				blockNum := sg.BlockNumber()
				rec := sg.GetTransaction(tx.SwapTx)

				if blockNum == 0 || rec == nil {
					span.Finish()
					continue
				}

				confirmationBlock := blockNum - rec.BlockNum
				if confirmationBlock == 0 {
					span.Finish()
					continue
				}
				_ = tx.UpdateSwapTx(spanCtx, int(confirmationBlock), chain)
			}
			span.Finish()
			t.Reset(time.Second * 5)
		}
	}

}
