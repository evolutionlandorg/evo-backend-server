package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
)

type LandData struct {
	ID         uint `gorm:"primary_key"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	TokenId    string `json:"token_id"`
	IsReserved int    `json:"is_reserved"` // 1 保留地 2 抽奖地块 3 宝箱
	ZAxis      int    `json:"z_axis"`
	GoldRate   int    `json:"gold_rate"`
	WoodRate   int    `json:"wood_rate"`
	WaterRate  int    `json:"water_rate"`
	FireRate   int    `json:"fire_rate"`
	SoilRate   int    `json:"soil_rate"`
	IsSpecial  int    `json:"is_special"`
	HasBox     int    `json:"has_box"`
	LandId     uint   `json:"land_id"`
}

type LandDataJson struct {
	IsReserved int `json:"is_reserved"`
	GoldRate   int `json:"gold_rate"`
	WoodRate   int `json:"wood_rate"`
	WaterRate  int `json:"water_rate"`
	FireRate   int `json:"fire_rate"`
	SoilRate   int `json:"soil_rate"`
	IsSpecial  int `json:"is_special"`
	HasBox     int `json:"has_box"`
}

func (ld *LandData) New(ctx context.Context, tx ...*gorm.DB) error {
	if len(tx) != 0 && tx[0] != nil {
		return tx[0].Create(&ld).Error
	}
	return util.WithContextDb(ctx).Create(&ld).Error
}

func GetLandData(tokenId string) (*LandData, error) {
	chain := GetChainByTokenId(tokenId)

	var ld LandData

	sg := storage.New(chain)
	result, err := sg.GetResourceRateAttr(tokenId)
	if err != nil {
		return &ld, err
	}
	if len(result) == 0 {
		return &ld, fmt.Errorf("get %s land resource error: result is empty", tokenId)
	}

	resource := util.ResourceDecode(result)
	ld.TokenId = tokenId
	ld.GoldRate = resource[0]
	ld.WoodRate = resource[1]
	ld.WaterRate = resource[2]
	ld.FireRate = resource[3]
	ld.SoilRate = resource[4]
	isReserved, _ := sg.LandMask(tokenId)
	ld.IsReserved = isReserved
	if ld.IsReserved == 4 {
		ld.HasBox = 1
	}
	return &ld, nil
}

func landResources(ctx context.Context, tokenId string) LandDataJson {
	db := util.WithContextDb(ctx)
	var ldj LandDataJson
	cacheKey := fmt.Sprintf("evo:Resources:%s", tokenId)
	if cache := util.GetCache(ctx, cacheKey); cache != nil {
		_ = json.Unmarshal(cache, &ldj)
	} else {
		query := db.Table("land_data").Where("token_id = ?", tokenId).First(&ldj)
		if query.RecordNotFound() {
			return ldj
		}
		cache, _ = json.Marshal(ldj)
		_ = util.SetCache(ctx, cacheKey, cache, 3600)
	}
	return ldj
}

func refreshLandData(ctx context.Context, tokenId string) error {
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	land := GetLandByTokenId(ctx, tokenId)
	if land == nil {
		return errors.New("land not find")
	}
	newLandData, err := GetLandData(tokenId)
	if err != nil {
		db.DbRollback()
		return errors.New("getLandData error")
	}
	db.Table("land_data").Where("token_id = ?", tokenId).UpdateColumn(map[string]interface{}{
		"is_reserved": newLandData.IsReserved,
		"gold_rate":   newLandData.GoldRate,
		"wood_rate":   newLandData.WoodRate,
		"water_rate":  newLandData.WaterRate,
		"fire_rate":   newLandData.FireRate,
		"soil_rate":   newLandData.SoilRate,
		"has_box":     newLandData.HasBox,
	})
	db.DbCommit()
	cacheKey := fmt.Sprintf("evo:Resources:%s", tokenId)
	util.DelCache(ctx, cacheKey)
	return nil
}
