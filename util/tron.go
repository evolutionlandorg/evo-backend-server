package util

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/evolutionlandorg/block-scan/util"
	"go.uber.org/atomic"
)

var (
	trxFilterContractAddrs []string
	TronRpc                *TrxRPC
)

type (
	logger interface {
		Println(v ...interface{})
	}
	TrxRPC struct {
		Url    string
		Client TronGridClient
		Log    logger
	}
	TronGridClient struct {
		Client  *http.Client
		keys    []string
		useKeys *atomic.Int64
	}
)

func newTrxRPC(url string) *TrxRPC {
	var keys []string
	for _, v := range strings.Split(util.GetEnv("TRON-PRO-API-KEY", ""), ",") {
		if v == "" {
			continue
		}
		keys = append(keys, v)
	}

	return &TrxRPC{
		Url: url,
		Client: TronGridClient{Client: http.DefaultClient, keys: keys,
			useKeys: atomic.NewInt64(1)},
		Log: log.New(os.Stderr, "", log.LstdFlags),
	}
}

func InitTron() {
	trxFilterContractAddrs = GetContracts("Tron")
	TronRpc = newTrxRPC(Evo.RpcEndpoint["Tron"])
}

func IsFilterTrxContractByAddr(addr string) bool {
	return StringInSlice(addr, trxFilterContractAddrs)
}

func (t *TronGridClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	client := t.Client
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", contentType)
	index := t.useKeys.Add(1)
	req.Header.Set("TRON-PRO-API-KEY", t.keys[int(index)%len(t.keys)])
	var (
		resp *http.Response
		err  error
	)
	err = Try(func() error {
		resp, err = client.Do(req)
		return err
	}, 5)
	return resp, err
}
