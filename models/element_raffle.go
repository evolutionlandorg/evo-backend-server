package models

import (
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/evolutionlandorg/block-scan/services"
	evoServices "github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/jinzhu/gorm"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"golang.org/x/net/context"
)

type ElementRaffle struct {
	gorm.Model
	Tx           string `json:"tx" gorm:"type:varchar(124)"`
	Owner        string `json:"owner" gorm:"type:varchar(68)"`
	Chain        string `json:"chain" gorm:"type:varchar(24)"`
	Element      string `json:"element" gorm:"type:varchar(24)"`
	Timestamp    int64  `json:"timestamp"`
	IsWin        bool   `json:"is_win"`
	DrawMethod   string `json:"draw_method" gorm:"type:varchar(24)"`
	PrizeType    string `json:"prize_type"`
	PrizeTokenId string `json:"prize_token_id"`
}

func (ec *EthTransactionCallback) GoldRaffleCallback(ctx context.Context) error {
	return ec.raffleCallback(ctx, currencyGold)
}

func (ec *EthTransactionCallback) WoodRaffleCallback(ctx context.Context) error {
	return ec.raffleCallback(ctx, currencyWood)
}

func (ec *EthTransactionCallback) WaterRaffleCallback(ctx context.Context) error {
	return ec.raffleCallback(ctx, currencyWater)
}

func (ec *EthTransactionCallback) FireRaffleCallback(ctx context.Context) error {
	return ec.raffleCallback(ctx, currencyFire)
}

func (ec *EthTransactionCallback) SoilRaffleCallback(ctx context.Context) error {
	return ec.raffleCallback(ctx, currencySoil)
}

func (ec *EthTransactionCallback) raffleCallback(ctx context.Context, element string) error {
	if getTransactionDeal(ctx, ec.Tx, "raffle") != nil {
		return errors.New("tx exist")
	}
	txn := util.DbBegin(ctx)
	defer txn.DbRollback()
	chain := ec.Receipt.ChainSource
	address := []string{
		strings.ToLower(util.GetContractAddress("GoldRaffle", chain)),
		strings.ToLower(util.GetContractAddress("WoodRaffle", chain)),
		strings.ToLower(util.GetContractAddress("WaterRaffle", chain)),
		strings.ToLower(util.GetContractAddress("FireRaffle", chain)),
		strings.ToLower(util.GetContractAddress("SoilRaffle", chain)),
		strings.ToLower(util.GetContractAddress("objectOwnership", chain)),
	}
	var waitInsertData = make(map[string]*ElementRaffle)
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) == 0 || !util.StringInSlice(strings.ToLower(util.AddHex(log.Address, chain)), address) {
			continue
		}
		var drawMethod string
		switch util.AddHex(log.Topics[0]) {
		case evoServices.AbiEncodingMethod("Transfer(address,address,uint256)"):
			to := util.AddHex(util.TrimHex(log.Topics[2])[24:64], chain)
			var tokenId string
			if len(log.Topics) < 4 {
				tokenId = util.TrimHex(log.Data)
			} else {
				tokenId = util.TrimHex(log.Topics[3])
			}
			if v, ok := waitInsertData[strings.ToLower(to)]; ok && v.IsWin {
				waitInsertData[strings.ToLower(to)].PrizeTokenId = tokenId
				continue
			}
			waitInsertData[strings.ToLower(to)] = &ElementRaffle{PrizeTokenId: tokenId}
			continue
		case evoServices.AbiEncodingMethod("LargeDraw(address,uint256,uint8)"):
			drawMethod = "Large"
		case evoServices.AbiEncodingMethod("SmallDraw(address,uint256,uint8)"):
			drawMethod = "Small"
		default:
			continue
		}
		logSlice := util.LogAnalysis(log.Data)
		prizeType := util.U256(logSlice[2]).Int64()
		e := &ElementRaffle{
			Owner:      util.AddHex(util.TrimHex(logSlice[0])[24:64], chain),
			Chain:      chain,
			Element:    element,
			Timestamp:  ec.BlockTimestamp,
			Tx:         ec.Tx,
			DrawMethod: drawMethod,
			IsWin:      prizeType != 0,
			PrizeType:  GetAssetTypeById(int(prizeType)),
		}
		if v, ok := waitInsertData[strings.ToLower(e.Owner)]; ok && v.PrizeTokenId != "" && e.IsWin {
			e.PrizeTokenId = v.PrizeTokenId
		}
		waitInsertData[strings.ToLower(e.Owner)] = e
	}
	for _, v := range waitInsertData {
		txn.Create(v)
	}
	if err := ec.NewUniqueTransaction(txn, "raffle"); err != nil {
		return err
	}

	txn.DbCommit()
	return txn.Error
}

type filterLogOpt struct {
	startBlock uint64
	client     *ethclient.Client
	parseFunc  func(*EthTransactionCallback)
	contracts  []string
	chain      string
}

func filterLogs(ctx context.Context, opt filterLogOpt) uint64 {
	query := new(ethereum.FilterQuery)
	for _, v := range opt.contracts {
		query.Addresses = append(query.Addresses, common.HexToAddress(v))
	}
	startBlock := opt.startBlock
	for {
		endBlock, _ := opt.client.BlockNumber(context.Background())
		if endBlock == 0 {
			time.Sleep(time.Second)
			continue
		}
		if endBlock <= startBlock {
			return endBlock
		}

		if endBlock-startBlock >= 500 {
			endBlock = startBlock + 500
		}

		query.FromBlock = big.NewInt(int64(startBlock))
		query.ToBlock = big.NewInt(int64(endBlock))
		rawLogs, err := opt.client.FilterLogs(ctx, *query)

		if err != nil {
			log.Warn("%s FilterLogs error: %v. trying again.", opt.chain, err)
			continue
		}

		var (
			data        = make(map[string]*EthTransactionCallback)
			blockNumber = make(map[uint64]uint64)
		)

		for _, v := range rawLogs {
			tx := v.TxHash.Hex()
			if _, ok := blockNumber[v.BlockNumber]; !ok {
				result, err := util.TryReturn(func() (result interface{}, err error) {
					return opt.client.BlockByNumber(ctx, big.NewInt(int64(v.BlockNumber)))
				}, 10)
				if err != nil {
					log.Error("%s %s get block by number error: %s", opt.chain, tx, err)
					continue
				}
				blockNumber[v.BlockNumber] = result.(*types.Block).Time()
			}
			if _, ok := data[tx]; !ok {
				data[tx] = &EthTransactionCallback{
					Tx:             v.TxHash.Hex(),
					BlockTimestamp: int64(blockNumber[v.BlockNumber]),
				}
			}
		}

		for _, v := range data {
			result, err := util.TryReturn(func() (result interface{}, err error) {
				resp, err := storage.New(opt.chain).ReceiptLog(v.Tx)
				if err != nil {
					time.Sleep(time.Second)
					return nil, err
				}
				return resp, nil
			}, 10)
			util.Panic(err)
			data[v.Tx].Receipt = result.(*services.Receipts)
			opt.parseFunc(data[v.Tx])
		}
		log.Warn("%s %d-%d block high filter logs %d", opt.chain, startBlock, endBlock, len(data))
		startBlock = endBlock
	}
}

func RefreshElementRaffle(ctx context.Context, chain []string, startBlock []int64) error {
	var wait sync.WaitGroup
	for index, v := range chain {
		wait.Add(1)
		go func(chain string, startBlock int64) {
			defer wait.Done()
			client, err := ethclient.DialContext(ctx, storage.GetChainWssRpc(chain))
			util.Panic(err)
			contracts := []string{
				strings.ToLower(util.GetContractAddress("GoldRaffle", chain)),
				strings.ToLower(util.GetContractAddress("WoodRaffle", chain)),
				strings.ToLower(util.GetContractAddress("WaterRaffle", chain)),
				strings.ToLower(util.GetContractAddress("FireRaffle", chain)),
				strings.ToLower(util.GetContractAddress("SoilRaffle", chain)),
			}
			filterLogs(ctx, filterLogOpt{
				startBlock: uint64(startBlock),
				client:     client,
				contracts:  contracts,
				chain:      chain,
				parseFunc: func(ec *EthTransactionCallback) {
					var f func(ctx context.Context) error
					for _, l := range ec.Receipt.Logs {
						if !util.StringInSlice(strings.ToLower(l.Address), contracts) {
							continue
						}
						switch strings.ToLower(l.Address) {
						case strings.ToLower(util.GetContractAddress("GoldRaffle", chain)):
							f = ec.GoldRaffleCallback
						case strings.ToLower(util.GetContractAddress("WoodRaffle", chain)):
							f = ec.WoodRaffleCallback
						case strings.ToLower(util.GetContractAddress("WaterRaffle", chain)):
							f = ec.WaterRaffleCallback
						case strings.ToLower(util.GetContractAddress("FireRaffle", chain)):
							f = ec.FireRaffleCallback
						case strings.ToLower(util.GetContractAddress("SoilRaffle", chain)):
							f = ec.SoilRaffleCallback
						default:
							continue
						}
					}
					ts := &TransactionScan{
						Tx:             ec.Tx,
						Chain:          chain,
						BlockNumber:    cast.ToInt64(ec.Receipt.BlockNumber),
						BlockTimestamp: int64(ec.BlockTimestamp),
					}
					data, _ := jsoniter.Marshal(ec.Receipt.Logs)
					ts.Logs = string(data)
					ts.New(context.TODO())
					if f != nil {
						_ = f(context.TODO())
					}
				},
			})
		}(v, startBlock[index])
	}
	wait.Wait()
	return nil
}
