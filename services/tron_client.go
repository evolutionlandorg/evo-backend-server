package services

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strings"

	"github.com/evolutionlandorg/block-scan/services"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/shopspring/decimal"
)

type (
	TrxBlock struct {
		BlockID     string `json:"blockID"`
		BlockHeader struct {
			RawData struct {
				Number         int    `json:"number"`
				ParentHash     string `json:"parentHash"`
				Timestamp      uint64 `json:"timestamp"`
				TxTrieRoot     string `json:"txTrieRoot"`
				Version        int    `json:"version"`
				WitnessAddress string `json:"witness_address"`
			} `json:"raw_data"`
			WitnessSignature string `json:"witness_signature"`
		} `json:"block_header"`
		Transactions []struct {
			RawData struct {
				Contract []struct {
					Parameter struct {
						TypeURL string `json:"type_url"`
						Value   struct {
							ContractAddress string `json:"contract_address"`
							Data            string `json:"data"`
							OwnerAddress    string `json:"owner_address"`
						} `json:"value"`
					} `json:"parameter"`
					Type string `json:"type"`
				} `json:"contract"`
				Expiration    int    `json:"expiration"`
				RefBlockBytes string `json:"ref_block_bytes"`
				RefBlockHash  string `json:"ref_block_hash"`
			} `json:"raw_data"`
			Ret []struct {
				ContractRet string `json:"contractRet"`
			} `json:"ret"`
			Signature []string `json:"signature"`
			TxID      string   `json:"txID"`
		} `json:"transactions"`
	}

	TrxTransaction struct {
		ContractAddress string `json:"contract_address"`
		RawData         struct {
			Contract []struct {
				Parameter struct {
					TypeURL string `json:"type_url"`
					Value   struct {
						NewContract struct {
							Abi struct {
								Entrys []struct {
									Inputs []struct {
										Name string `json:"name"`
										Type string `json:"type"`
									} `json:"inputs"`
									Name    string `json:"name"`
									Outputs []struct {
										Name string `json:"name"`
										Type string `json:"type"`
									} `json:"outputs"`
									StateMutability string `json:"stateMutability"`
									Type            string `json:"type"`
								} `json:"entrys"`
							} `json:"abi"`
							Bytecode                   string `json:"bytecode"`
							ConsumeUserResourcePercent int    `json:"consume_user_resource_percent"`
							Name                       string `json:"name"`
							OriginAddress              string `json:"origin_address"`
						} `json:"new_contract"`
						OwnerAddress string `json:"owner_address"`
					} `json:"value"`
				} `json:"parameter"`
				Type string `json:"type"`
			} `json:"contract"`
			Expiration    int    `json:"expiration"`
			FeeLimit      int    `json:"fee_limit"`
			RefBlockBytes string `json:"ref_block_bytes"`
			RefBlockHash  string `json:"ref_block_hash"`
			Timestamp     int    `json:"timestamp"`
		} `json:"raw_data"`
		Ret []struct {
			ContractRet string `json:"contractRet"`
		} `json:"ret"`
		Signature []string `json:"signature"`
		TxID      string   `json:"txID"`
	}

	TrxTxReceipt struct {
		BlockNumber          int      `json:"blockNumber"`
		BlockTimeStamp       int      `json:"blockTimeStamp"`
		ContractResult       []string `json:"contractResult"`
		ContractAddress      string   `json:"contract_address"`
		Fee                  int      `json:"fee"`
		ID                   string   `json:"id"`
		InternalTransactions []struct {
			CallValueInfo     []struct{} `json:"callValueInfo"`
			CallerAddress     string     `json:"caller_address"`
			Hash              string     `json:"hash"`
			TransferToAddress string     `json:"transferTo_address"`
		} `json:"internal_transactions"`
		Log []struct {
			Address string   `json:"address"`
			Data    string   `json:"data"`
			Topics  []string `json:"topics"`
		} `json:"log"`
		Receipt struct {
			EnergyFee        int    `json:"energy_fee"`
			EnergyUsageTotal int    `json:"energy_usage_total"`
			NetUsage         int    `json:"net_usage"`
			Result           string `json:"result"`
		} `json:"receipt"`
	}

	TrxContractExecution struct {
		ConstantResult []string `json:"constant_result"`
		Result         struct {
			Message string `json:"message"`
			Result  bool   `json:"result"`
		} `json:"result"`
		Transaction struct {
			RawData struct {
				Contract []struct {
					Parameter struct {
						TypeURL string `json:"type_url"`
						Value   struct {
							CallValue       int    `json:"call_value"`
							ContractAddress string `json:"contract_address"`
							Data            string `json:"data"`
							OwnerAddress    string `json:"owner_address"`
						} `json:"value"`
					} `json:"parameter"`
					Type string `json:"type"`
				} `json:"contract"`
				Expiration    int    `json:"expiration"`
				FeeLimit      int    `json:"fee_limit"`
				RefBlockBytes string `json:"ref_block_bytes"`
				RefBlockHash  string `json:"ref_block_hash"`
				Timestamp     int    `json:"timestamp"`
			} `json:"raw_data"`
			Ret []struct {
				Ret string `json:"ret"`
			} `json:"ret"`
			TxID string `json:"txID"`
		} `json:"transaction"`
		Txid string `json:"txid"`
	}
)

type TronClient struct {
	*util.TrxRPC
}

// Balance of returns the account of address TRX10 balance
func (rpc *TronClient) Balance(address string) *big.Int {
	url := rpc.Url + "/walletsolidity/getaccount"
	body := fmt.Sprintf("{\"address\":\"%s\"}", address)
	response, err := rpc.Client.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return big.NewInt(0)
	}
	data, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		return big.NewInt(0)
	}
	resp := struct {
		Address string `json:"address"`
		Balance int64  `json:"balance"`
	}{}
	if err := json.Unmarshal(data, &resp); err != nil {
		return big.NewInt(0)
	}
	return big.NewInt(resp.Balance)
}

// BlockNumber returns the TRX current block number
// return the block number with format in uint64
func (rpc *TronClient) BlockNumber() uint64 {
	url := rpc.Url + "/walletsolidity/getnowblock"
	response, err := rpc.Client.Post(url, "application/json", nil)
	if err != nil {
		return 0
	}

	data, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		return 0
	}
	resp := struct {
		BlockID     string `json:"blockID"`
		BlockHeader struct {
			RawData struct {
				Number         int    `json:"number"`
				ParentHash     string `json:"parentHash"`
				Timestamp      int    `json:"timestamp"`
				TxTrieRoot     string `json:"txTrieRoot"`
				Version        int    `json:"version"`
				WitnessAddress string `json:"witness_address"`
			} `json:"raw_data"`
			WitnessSignature string `json:"witness_signature"`
		} `json:"block_header"`
	}{}

	if err := json.Unmarshal(data, &resp); err != nil {
		// rpc.log.Println(err)
		return 0
	}

	return uint64(resp.BlockHeader.RawData.Number)

}

// GetBlockByNumber returns the TRX block data
// @number is block number
// return the block data point
func (rpc *TronClient) GetBlockByNumber(number uint64) *TrxBlock {
	url := rpc.Url + "/walletsolidity/getblockbynum"
	body := fmt.Sprintf("{\"num\":%d}", number)

	response, err := rpc.Client.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return nil
	}

	data, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		return nil
	}
	resp := new(TrxBlock)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil
	}

	return resp
}

// GetSolidityTransactionByHash returns the TRX transaction data
func (rpc *TronClient) GetSolidityTransactionByHash(id string) *TrxTransaction {

	url := rpc.Url + "/walletsolidity/gettransactionbyid"
	body := fmt.Sprintf("{\"value\":\"%s\"}", id)

	response, err := rpc.Client.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return nil
	}

	data, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		return nil
	}

	resp := new(TrxTransaction)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil
	}

	return resp
}

// GetTxReceiptByHash returns the TRX transaction receipt data
func (rpc *TronClient) GetTxReceiptByHash(id string) *TrxTxReceipt {

	url := rpc.Url + "/walletsolidity/gettransactioninfobyid"
	body := fmt.Sprintf("{\"value\":\"%s\"}", id)

	response, err := rpc.Client.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return nil
	}

	data, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		return nil
	}
	resp := new(TrxTxReceipt)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil
	}

	return resp
}

func (rpc *TronClient) ContractCall(contractAddress, _ string, feeLimit, callValue uint64, function string, params ...string) *TrxContractExecution {

	url := rpc.Url + "/wallet/triggersmartcontract"
	body := fmt.Sprintf("{\"contract_address\":\"%s\",\"function_selector\":\"%s\",\"fee_limit\":%d,\"call_value\":%d,\"owner_address\":\"%s\",\"parameter\":\"",
		contractAddress,
		function,
		feeLimit,
		callValue,
		contractAddress)

	for _, p := range params {
		body += util.Padding(p)
	}

	body += "\"}"

	response, err := rpc.Client.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return nil
	}

	data, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		log.Error("tron ReadAll error: %s", err.Error())
		return nil
	}

	resp := new(TrxContractExecution)
	if err := json.Unmarshal(data, resp); err != nil {
		log.Error("tron Unmarshal error: %s", err.Error())
		return nil
	}

	return resp
}

type TronRawTransaction struct {
	Signature []string `json:"signature"`
	TxID      string   `json:"tx_id"`
	RawData   RawData  `json:"raw_data"`
}
type RawData struct {
	Contract []struct {
		Parameter struct {
			TypeURL string                 `json:"type_url"`
			Value   map[string]interface{} `json:"value"`
		} `json:"parameter"`
		Type string `json:"type"`
	} `json:"contract"`
	Expiration    int    `json:"expiration"`
	Timestamp     int    `json:"timestamp"`
	RefBlockBytes string `json:"ref_block_bytes"`
	RefBlockHash  string `json:"ref_block_hash"`
}

type RawTransactionReturn struct {
	Code   string `json:"code"`
	Result bool   `json:"result"`
}

func (rpc *TronClient) SendRawTransaction(_ string, transaction []byte) *RawTransactionReturn {
	url := rpc.Url + "/wallet/broadcasttransaction"
	response, err := rpc.Client.Post(url, "application/json", strings.NewReader(string(transaction)))
	if err != nil {
		return nil
	}
	data, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		return nil
	}
	var re RawTransactionReturn
	_ = json.Unmarshal(data, &re)
	return &re
}

type TransactionParam struct {
	ToAddress    string `json:"to_address"`
	OwnerAddress string `json:"owner_address"`
	Amount       string `json:"amount"`
}

func (rpc *TronClient) CreateTransaction(toAddress, ownerAddress string, amount decimal.Decimal) string {
	url := rpc.Url + "/wallet/createtransaction"
	body := TransactionParam{ToAddress: toAddress, OwnerAddress: ownerAddress, Amount: amount.String()}
	b, _ := json.Marshal(body)
	response, err := rpc.Client.Post(url, "application/json", strings.NewReader(string(b)))
	if err != nil {
		return ""
	}
	data, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		return ""
	}
	resp := new(TrxContractExecution)
	log.Debug("TrxContractExecution url=%s", string(data))
	if err := json.Unmarshal(data, resp); err != nil {
		return ""
	}
	return string(data)
}

type TransactionEvent struct {
	BlockNumber     int                    `json:"block_number"`
	BlockTimestamp  int                    `json:"block_timestamp"`
	ContractAddress string                 `json:"contract_address"`
	EventIndex      int                    `json:"event_index"`
	EventName       string                 `json:"event_name"`
	Result          map[string]interface{} `json:"result"`
	ResultType      map[string]interface{} `json:"result_type"`
	TransactionId   string                 `json:"transaction_id"`
	ResourceNode    string                 `json:"resource_node"`
}

func (rpc *TronClient) GetTransactionEvent(tx string) *[]TransactionEvent {
	url := rpc.Url + "/event/transaction/" + tx
	body := util.HttpGet(url)
	var transactionEvent []TransactionEvent
	_ = json.Unmarshal(body, &transactionEvent)
	return &transactionEvent
}

func (rpc *TronClient) GetContractEvent(address string) *[]TransactionEvent {
	url := fmt.Sprintf("%s/event/contract/%s?&size=%d", rpc.Url, address, 10)
	body := util.HttpGet(url)
	var transactionEvent []TransactionEvent
	_ = json.Unmarshal(body, &transactionEvent)
	return &transactionEvent
}

type SampleTransaction struct {
	Ret []struct {
		ContractRet string `json:"contractRet"`
	} `json:"ret"`
}

func (rpc *TronClient) GetTransactionByHash(id string) *SampleTransaction {
	url := rpc.Url + "/wallet/gettransactionbyid"
	body := fmt.Sprintf("{\"value\":\"%s\"}", id)
	response, err := rpc.Client.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return nil
	}
	log.Debug("GetTransactionByHash url=%s", url)
	data, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		return nil
	}
	resp := new(SampleTransaction)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil
	}

	return resp
}

func BuildReceipt(txp *TrxTxReceipt) *services.Receipts {
	var ret services.Receipts
	ret.BlockNumber = util.AddHex(fmt.Sprintf("%x", txp.BlockNumber))
	if txp.Receipt.Result == "SUCCESS" {
		ret.Status = "0x1"
	} else {
		ret.Status = "0x0"
	}
	logs := make([]services.Log, 0)
	for _, v := range txp.Log {
		logs = append(logs, services.Log{Topics: v.Topics, Data: v.Data, Address: util.AddHex(v.Address, "Tron")})
	}
	ret.Logs = logs
	ret.ChainSource = "Tron"
	ret.Solidity = true
	return &ret
}

func (rpc *TronClient) BuildReceiptFromTronEvent(tx string) *Receipts {
	eventLog := rpc.GetTransactionEvent(tx)
	if len(*eventLog) == 0 {
		return nil
	}
	return rpc.DelEvent(eventLog)
}

func (rpc *TronClient) DelEvent(eventLog *[]TransactionEvent) *Receipts {
	var ret Receipts
	logs := make([]Log, 0)
	for _, v := range *eventLog {
		contractAddress := util.AddHex(util.TrxBase58toHexAddress(v.ContractAddress), "Tron")
		if !util.IsFilterTrxContractByAddr(contractAddress) {
			continue
		}
		ret.BlockNumber = util.AddHex(fmt.Sprintf("%x", v.BlockNumber))

		var (
			typeArr []string
			dataArr []string
		)
		abi, err := util.AnalysisAbiEvent(contractAddress, "Tron")
		if err != nil {
			continue
		}
		eventInstant := util.GetAbiEventInputs(abi, v.EventName)
		for _, paramType := range eventInstant {
			typeArr = append(typeArr, paramType.Type)
		}
		eventName := AbiEncodingMethod(v.EventName + "(" + strings.Join(typeArr, ",") + ")")
		topics := []string{eventName}
		for _, paramType := range eventInstant {
			paramData := v.Result[paramType.Name]
			param := ""
			switch v.ResultType[paramType.Name] {
			case "uint256", "uint8":
				param = util.DecodeInputU256(decimal.RequireFromString(paramData.(string)).Coefficient())
			case "address":
				param = util.Padding(util.TrimHex(paramData.(string)))
			case "dynamicbytes":
				param = paramData.(string)
			case "bool":
				paramStr := "0"
				s := reflect.ValueOf(paramData)
				if s.Kind() == reflect.String {
					if paramData.(string) == "true" {
						paramStr = "1"
					}
				} else if s.Kind() == reflect.Bool {
					if paramData.(bool) {
						paramStr = "1"
					}
				}
				param = util.Padding(paramStr)
			}
			if param != "" {
				if paramType.Indexed {
					topics = append(topics, util.AddHex(param))
				} else {
					dataArr = append(dataArr, param)
				}
			}

		}
		logs = append(logs, Log{
			Topics:  topics,
			Data:    util.AddHex(strings.Join(dataArr, "")),
			Address: contractAddress,
		})
	}
	ret.Logs = logs
	ret.Status = "0x1"
	ret.Solidity = false
	ret.ChainSource = "Tron"
	return &ret
}
