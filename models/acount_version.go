package models

import (
	"context"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type AccountVersion struct {
	gorm.Model
	MemberId  uint            `json:"member_id"`
	AccountId uint            `json:"account_id"`
	Balance   decimal.Decimal `json:"balance" sql:"type:decimal(32,16);"`
	Locked    decimal.Decimal `json:"locked" sql:"type:decimal(32,16);"`
	Reason    string          `json:"reason"`
	Remark    string          `json:"remark"`
	Operator  uint            `json:"operator"`
	Currency  string          `json:"currency"`
}

type AccountVersionJson struct {
	CreatedAt time.Time       `json:"created_at"`
	Balance   decimal.Decimal `json:"balance" sql:"type:decimal(32,16);"`
	Reason    string          `json:"reason"`
	Remark    string          `json:"remark"`
	Currency  string          `json:"currency"`
}

type AvQuery struct {
	WhereQuery VersionQuery
	Row        int
	Page       int
}

type VersionQuery struct {
	Reason    string `json:"reason" table_name:"account_versions"`
	AccountId uint   `json:"account_id" table_name:"account_versions"`
	Remark    string `json:"remark" table_name:"account_versions"`
}

func (av *AccountVersion) New(db *util.GormDB) error {
	result := db.Create(&av)
	return result.Error
}

func (avq *AvQuery) GetHistory(ctx context.Context, AccountId []uint) ([]AccountVersionJson, int) {
	if len(AccountId) == 0 {
		return nil, 0
	}
	db := util.WithContextDb(ctx).Table("account_versions")
	var avs []AccountVersionJson
	var count int
	wheres, values := util.StructToSql(avq.WhereQuery)
	if len(wheres) != 0 {
		db = db.Where(strings.Join(wheres, "AND"), values...)
	}
	query := db.Where("account_id in (?)", AccountId).
		Offset(avq.Page * avq.Row).Limit(avq.Row).
		Order("id desc").
		Scan(&avs)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil, 0
	}
	db = db.Table("account_versions")
	if len(wheres) != 0 {
		db = db.Where(strings.Join(wheres, "AND"), values...)
	}
	db.Where(avq.WhereQuery).Where("account_id in (?)", AccountId).
		Count(&count)
	return avs, count
}
