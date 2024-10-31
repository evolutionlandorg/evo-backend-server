package daemons

import (
	"context"
	"crypto/md5"
	"fmt"
	"reflect"
	"strings"

	"github.com/evolutionlandorg/block-scan/services"
	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/itering/go-workers"
	"github.com/spf13/cast"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func RunWorker() {
	processCount := cast.ToInt(util.GetEnv("WORKER_PROCESS_COUNT", "10"))
	if processCount == 0 {
		processCount = 10
	}
	workers.Process("ethProcess", ethProcess, processCount)
	workers.Process("tronProcess", tronProcess, processCount)
	workers.Process("crabProcess", crabProcess, processCount)
	workers.Process("hecoProcess", crabProcess, processCount)
	workers.Process("bscProcess", crabProcess, processCount)
	workers.Process("polygonProcess", crabProcess, processCount)
	workers.Run()
}

type ChainPayload struct {
	Tx             string             `json:"tx"`
	Chain          string             `json:"chain"`
	ContractName   string             `json:"contract_name"`
	BlockTimestamp int64              `json:"block_timestamp"`
	Receipts       *services.Receipts `json:"receipts"`
}

func ethProcess(m *workers.Msg) {
	var payload ChainPayload
	args := m.Get("args")
	util.UnmarshalAny(&payload, args)

	ecInstant := &models.EthTransactionCallback{Tx: payload.Tx, Receipt: payload.Receipts, BlockTimestamp: payload.BlockTimestamp}
	wReflect := reflect.ValueOf(&ecInstant).Elem()
	deal := func() {
		methodName := fmt.Sprintf("%sCallback", payload.ContractName)
		span, ctx := tracer.StartSpanFromContext(context.TODO(), "daemons.worker",
			tracer.ServiceName("evo-backend-worker"),
			tracer.SpanType(ext.SpanTypeMessageConsumer),
			tracer.Measured(),
			tracer.Tag("worker-name", methodName),
		)
		defer span.Finish()
		methodFunc := wReflect.MethodByName(methodName)
		if !methodFunc.IsValid() {
			log.Warn("%s not found method %s", payload.Chain, methodName)
			return
		}
		log.Debug("%s %s block tx %s call %s", payload.Chain, payload.Receipts.BlockNumber, payload.Tx, methodName)
		res := methodFunc.Call([]reflect.Value{reflect.ValueOf(ctx)})
		if v := res[0].Interface(); v != nil {
			if err, ok := v.(error); ok && !strings.EqualFold(err.Error(), "tx exist") {
				log.Error("Process error %s. chain %s tx %s", err, payload.Chain, payload.Tx)
				util.WithContextDb(ctx).Create(&models.ParseTxError{
					Tx:        payload.Tx,
					Chain:     payload.Chain,
					Error:     err.Error(),
					ParseFunc: methodName,
					Receipts:  util.ToString(payload.Receipts),
				})
				return
			}
		}
	}

	key := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s", payload.Tx, payload.ContractName))))
	util.OnceTask(context.TODO(), fmt.Sprintf("ethProcess:%s", key), 5, deal)
}

func tronProcess(m *workers.Msg) {
	defer util.Recover("tronProcess error")
	ethProcess(m)
}

func crabProcess(m *workers.Msg) {
	defer util.Recover("crabProcess error")
	ethProcess(m)
}
