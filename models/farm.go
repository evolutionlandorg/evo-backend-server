// Package models provides ...
package models

import (
	"context"
	"time"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
)

type FarmAPR struct {
	gorm.Model
	Pool string `json:"pool"`
	Addr string `json:"addr"`
	APR  string `json:"apr"`
}

func RawAddFarmAPR(ctx context.Context, pool string, addr string, apr string) error {
	return util.WithContextDb(ctx).Create(&FarmAPR{
		Pool: pool,
		Addr: addr,
		APR:  apr,
	}).Error
}

// RemoveFarmAPRByTime 删除 invalidTime 之前的 address 记录
func RemoveFarmAPRByTime(ctx context.Context, addr string, invalidTime time.Time) error {
	db := util.WithContextDb(ctx)
	return db.Where("addr = ? AND created_at < ?", addr, invalidTime).Unscoped().Delete(new(FarmAPR)).Error
}

func RawFarmAPR(ctx context.Context, addr string) FarmAPR {
	var apr FarmAPR
	util.WithContextDb(ctx).Where("addr = ?", addr).Order("id desc").First(&apr)
	return apr
}
