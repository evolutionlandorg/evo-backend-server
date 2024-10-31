package models

import (
	"context"
	"testing"
	"time"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestRemoveFarmAPRByTime(t *testing.T) {
	assert.NoError(t, util.InitMysql(log.NewGormLog()))
	address := "0xtest-TestRemoveFarmAPRByTime"
	var deleteId []interface{}
	db := util.WithContextDb(context.TODO())
	now := time.Unix(1642435200, 0) // 2022-01-18 00:00:00
	for i := 5; i > 0; i-- {
		f := &FarmAPR{
			Model: gorm.Model{CreatedAt: now.AddDate(0, 0, -i)},
			Pool:  "test",
			Addr:  address,
			APR:   "0.1",
		}
		assert.NoError(t, db.Create(f).Error, "create test farm APR failed")
		deleteId = append(deleteId, f.ID)
	}
	defer func() {
		db.Model(new(FarmAPR)).Where("id IN (?)", deleteId).Unscoped().Delete(new(FarmAPR)).Limit(len(deleteId))
	}()

	assert.NoError(t, RemoveFarmAPRByTime(context.TODO(), address, now.AddDate(0, 0, -2)))
	var count int
	db.Where("addr = ?", address).Model(new(FarmAPR)).Count(&count)
	assert.Equal(t, count, 2)

	assert.NoError(t, RemoveFarmAPRByTime(context.TODO(), address, now.AddDate(0, 0, 1)))
	db.Where("addr = ?", address).Model(new(FarmAPR)).Count(&count)
	assert.Equal(t, count, 0)
}
