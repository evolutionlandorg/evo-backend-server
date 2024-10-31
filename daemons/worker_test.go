package daemons

import (
	"os"
	"testing"

	"github.com/evolutionlandorg/evo-backend/config"
	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/itering/go-workers"
	"github.com/stretchr/testify/assert"
)

func init() {

	util.Panic(util.InitMysql())
	util.Panic(os.Setenv("CONF_DIR", "../config"))
	config.InitApplication()
}

func Test_ethProcess(t *testing.T) {
	s := map[string]interface{}{"args": ChainPayload{
		Tx:           "0x4a513cc63d5852bca87282852446f4f0ede056fdcc9bb5324511dfdab3963cf4",
		Chain:        "Eth",
		ContractName: "LandResource",
	}}
	m, err := workers.NewMsg(util.ToString(s))
	assert.NoError(t, err)
	ethProcess(m)
}

func Test_CryptoKittiesTransfer(t *testing.T) {
	s := map[string]interface{}{"args": ChainPayload{
		Tx:           "0x580f02f91b5a993eab9318f5ef6b888bb475196928724389e9b97eb7310c4ce8",
		Chain:        "Eth",
		ContractName: "CryptoKitties",
	}}
	m, err := workers.NewMsg(util.ToString(s))
	assert.NoError(t, err)
	ethProcess(m)
}

func Test_GoldRushJoin(t *testing.T) {
	s := map[string]interface{}{"args": ChainPayload{
		Tx:           "0xe1b6c8ec4c6938ef505759e8eb1947960b57895d97c2c5df85b6ae76976a378a",
		Chain:        "Eth",
		ContractName: "Raffle",
	}}
	m, err := workers.NewMsg(util.ToString(s))
	assert.NoError(t, err)
	ethProcess(m)
}

func Test_PveTeam(t *testing.T) {
	// join
	s := map[string]interface{}{"args": ChainPayload{
		Tx:           "0x635e7b9000314fb6f9751aa092c54d22d358736b085c06382b40e6c277bc6ccb",
		Chain:        "Heco",
		ContractName: "ClockAuctionApostle",
	}}
	m, err := workers.NewMsg(util.ToString(s))
	assert.NoError(t, err)
	ethProcess(m)
	// exit
	// s = map[string]interface{}{"args": ChainPayload{
	// 	Tx:           "0xcd20d4816c92ea861440a3a61ec413deadc3e3dc2f2301c2e383699ac5f757a3",
	// 	Chain:        "Heco",
	// 	ContractName: "PveTeam",
	// }}
	// m, err = workers.NewMsg(util.ToString(s))
	// assert.NoError(t, err)
	// ethProcess(m)
}

func Test_apostlePro(t *testing.T) {
	// 转职
	m, _ := workers.NewMsg(util.ToString(map[string]interface{}{"args": ChainPayload{
		Tx:           "0x7d39985bb4ea7e8f1c461f5f787fe25f0b2bd85811e6946afbb41b46583af0e4",
		Chain:        "Heco",
		ContractName: "Apostle",
	}}))
	ethProcess(m)
	// 合成装备
	m, _ = workers.NewMsg(util.ToString(map[string]interface{}{"args": ChainPayload{
		Tx:           "0x4924243243e4be46e0672dedc23ce7b9e079fa0fc3d5142aae41e1b7c695251e",
		Chain:        "Heco",
		ContractName: "CraftBase",
	}}))
	ethProcess(m)
	// 装备武器
	m, _ = workers.NewMsg(util.ToString(map[string]interface{}{"args": ChainPayload{
		Tx:           "0xffad9e0f0aff91684263fa20af3ac6d14da14f1cd2a155251b3a0b65fb2387d9",
		Chain:        "Heco",
		ContractName: "Apostle",
	}}))
	ethProcess(m)
	// 卸下装备
	m, _ = workers.NewMsg(util.ToString(map[string]interface{}{"args": ChainPayload{
		Tx:           "0x32aa6c3858e78994d8c311ee1c7d707f03194371f0b868a1db439926186f5eb7",
		Chain:        "Heco",
		ContractName: "Apostle",
	}}))
	ethProcess(m)
	// 强化装备
	m, _ = workers.NewMsg(util.ToString(map[string]interface{}{"args": ChainPayload{
		Tx:           "0x06a804f75e20700056af45b74c23b241b69ac4b39a70a25f97d2d40738f8539a",
		Chain:        "Heco",
		ContractName: "CraftBase",
	}}))
	ethProcess(m)
}
