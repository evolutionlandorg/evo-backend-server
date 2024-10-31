package models

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
)

type KeyStore struct {
	gorm.Model
	UsedAt   int    `json:"used_at"`
	MemberId uint   `json:"member_id"`
	Key      string `json:"key"`
	KeyHash  string `json:"key_hash"`
}

func (ec *EthTransactionCallback) RolesUpdaterCallback(ctx context.Context) (err error) {
	if getTransactionDeal(ctx, ec.Tx, "keyStore") != nil {
		return errors.New("tx exist")
	}
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(log.Address, util.GetContractAddress("rolesUpdater")) {
			eventName := log.Topics[0]
			switch eventName {
			case services.AbiEncodingMethod("UpdateTesterRole(address,uint256,bytes32)"):
				logSlice := util.LogAnalysis(log.Data)
				address := util.AddHex(log.Topics[1][26:66])
				keyHash := util.AddHex(logSlice[0])
				member := GetMemberByAddress(ctx, address, EthChain)
				if member != nil {
					member.UploadPlayRoleByKey(ctx, db, keyHash)
				}
			}
		}
	}

	if err != nil {
		return err
	}

	if err := ec.NewUniqueTransaction(db, "keyStore"); err != nil {
		return err
	}

	db.DbCommit()
	return db.Error
}

func (ec *EthTransactionCallback) UserRolesCallback(ctx context.Context) (err error) {
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(log.Address, util.GetContractAddress("userRoles")) {
			// 忽略rolesUpdater 使用keystore 的情况
			if len(ec.Receipt.Logs) == 2 && strings.EqualFold(ec.Receipt.Logs[1].Address, util.GetContractAddress("rolesUpdater")) {
				break
			}
			eventName := log.Topics[0]
			switch eventName {
			case services.AbiEncodingMethod("RoleAdded(address,string)"):
				member := GetMemberByAddress(ctx, util.AddHex(log.Topics[1][26:66]), TronChain)
				if member != nil {
					member.PlayerRole = 1
					_ = member.updateField(ctx, db)
				}
			}
		}
	}
	db.DbCommit()
	return err
}

func useKeyStore(ctx context.Context, memberId uint, key string) bool {
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	query := db.Table("key_stores").Where("key_hash = ?", key).Where("used_at =?", 0).UpdateColumn(KeyStore{
		MemberId: memberId,
		UsedAt:   int(time.Now().Unix()),
	})
	db.DbCommit()
	return query.RowsAffected > 0
}

func ExamineKeyStore(ctx context.Context, key string, memberId uint) bool {
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	var k KeyStore
	query := db.Table("key_stores").Where("`key` = ?", key).Where("used_at =?", 0).First(&k)
	if query.RecordNotFound() {
		return false
	} else {
		if k.MemberId != 0 {
			return k.MemberId == memberId
		} else {
			db.Table("key_stores").Where("member_id = ?", memberId).UpdateColumn(map[string]interface{}{"member_id": 0})
			db.Model(&k).UpdateColumn(KeyStore{MemberId: memberId})
			db.DbCommit()
			return true
		}
	}
}
