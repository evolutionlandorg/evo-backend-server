package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/evolutionlandorg/evo-backend/util"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/shopspring/decimal"
)

var defiBoxConf = map[string]string{
	"sign_version": "1.0",
	"access_key":   util.GetEnv("defibox_access_key", ""),
	"secret_key":   util.GetEnv("secret_key", ""),
}

type VSCurrency struct {
	USD decimal.Decimal `json:"usd"`
}

type Price struct {
	DarwiniaNetworkNativeToken VSCurrency `json:"darwinia-network-native-token"`
}

type PriceKton struct {
	DarwiniaCommitmentToken VSCurrency `json:"darwinia-commitment-token"`
}

func GetRingPrice() (decimal.Decimal, error) {
	var price Price
	res := httpGet("https://api.coingecko.com/api/v3/simple/price?ids=darwinia-network-native-token&vs_currencies=usd")
	if err := json.Unmarshal([]byte(res), &price); err != nil {
		return decimal.NewFromInt(0), err
	}
	return price.DarwiniaNetworkNativeToken.USD, nil
}

func GetKtonPrice() (decimal.Decimal, error) {
	var price PriceKton
	res := httpGet("https://api.coingecko.com/api/v3/simple/price?ids=darwinia-commitment-token&vs_currencies=usd")
	if err := json.Unmarshal([]byte(res), &price); err != nil {
		return decimal.NewFromInt(0), err
	}
	return price.DarwiniaCommitmentToken.USD, nil
}

func ReportDefiBoxData(subUrl string, postData map[string]string) string {
	thisUrl := "https://www.defibox.com" + subUrl
	postData["signature"] = createDefiBoxSignature(postData, defiBoxConf)
	postData["sign_version"] = defiBoxConf["sign_version"]
	postData["access_key"] = defiBoxConf["access_key"]
	log.Debug("request ReportDefiBoxData. url=%s  body=%+v", thisUrl, postData)
	return httpPost(thisUrl, postData)
}

func createDefiBoxSignature(postData, _ map[string]string) string {
	sv := defiBoxConf["sign_version"]
	ak := defiBoxConf["access_key"]
	sk := defiBoxConf["secret_key"]
	ts := strconv.FormatInt(time.Now().UTC().UnixNano()/int64(time.Millisecond), 10)
	keys := make([]string, 0, 32)
	for k := range postData {
		keys = append(keys, k)
	}
	postData["timestamp"] = ts
	sort.Strings(keys)
	strList := make([]string, 0, 32)
	strList = append(strList, fmt.Sprintf("sign_version=%s", sv))
	strList = append(strList, fmt.Sprintf("access_key=%s", ak))
	strList = append(strList, fmt.Sprintf("secret_key=%s", sk))
	strList = append(strList, fmt.Sprintf("timestamp=%s", ts))
	for _, key := range keys {
		strList = append(strList, fmt.Sprintf("%s=%s", key, postData[key]))
	}
	sig := strings.Join(strList, "@")
	hash := sha256.New()
	_, _ = hash.Write([]byte(sig))
	bytes := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(bytes)
}

func httpGet(queryUrl string) string {
	u, _ := url.Parse(queryUrl)
	retStr, err := http.Get(u.String())
	if err != nil {
		return err.Error()
	}
	result, err := io.ReadAll(retStr.Body)
	_ = retStr.Body.Close()
	if err != nil {
		return err.Error()
	}
	return string(result)
}

func httpPost(queryUrl string, postData map[string]string) string {
	data, err := json.Marshal(postData)
	if err != nil {
		return err.Error()
	}

	body := bytes.NewBuffer([]byte(data))

	retStr, err := http.Post(queryUrl, "application/json;charset=utf-8", body)

	if err != nil {
		return err.Error()
	}
	result, err := io.ReadAll(retStr.Body)
	_ = retStr.Body.Close()
	if err != nil {
		return err.Error()
	}
	return string(result)
}
