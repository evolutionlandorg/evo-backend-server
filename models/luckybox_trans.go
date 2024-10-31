package models

import (
	"context"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
)

type LuckyboxTrans struct {
	gorm.Model
	Tx string `json:"tx"`
}

func (lt *LuckyboxTrans) New(ctx context.Context) error {
	result := util.WithContextDb(ctx).Create(&lt)
	return result.Error
}

func GetLt(ctx context.Context, tx string) *LuckyboxTrans {
	db := util.WithContextDb(ctx)
	lt := LuckyboxTrans{}
	query := db.Where("tx = ?", tx).First(&lt)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	return &lt
}
