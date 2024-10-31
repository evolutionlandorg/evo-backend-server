package models

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/jinzhu/gorm"
)

type MemberTakeBack struct {
	MemberId uint `json:"member_id" gorm:"primary_key;auto_increment:false"`

	CrabRing     int `json:"crab_ring"`
	CrabKton     int `json:"crab_kton"`
	CrabMaterial int `json:"crab_material" sql:"default:0"`

	HecoRing     int `json:"heco_ring"`
	HecoKton     int `json:"heco_kton"`
	HecoMaterial int `json:"heco_material" sql:"default:0"`

	PolygonRing     int `json:"polygon_ring"`
	PolygonKton     int `json:"polygon_kton"`
	PolygonMaterial int `json:"polygon_material" sql:"default:0"`
}

type OriTakeBackNonce struct {
	WithdrawNonce int `json:"withdraw_nonce"`
	TronNonce     int `json:"tron_nonce"`
	KtonNonce     int `json:"kton_nonce"`
	TronKtonNonce int `json:"tron_kton_nonce"`
}

func (m *Member) TakeBackNonce(ctx context.Context) *MemberTakeBack {
	var t MemberTakeBack
	db := util.WithContextDb(ctx)
	if query := db.FirstOrCreate(&t, MemberTakeBack{MemberId: m.ID}); query.Error != nil {
		return nil
	}
	return &t
}

func (m *Member) MaterialNonce(ctx context.Context, chain string) int {
	takeBack := m.TakeBackNonce(ctx)
	if takeBack == nil {
		return 0
	}
	nonceField := strings.ToLower(fmt.Sprintf("%s_%s", chain, "material"))
	if val, _ := util.GetFieldValByTag(nonceField, "json", takeBack); val != nil {
		return val.(int)
	}
	return 0
}

func (m *Member) UpdateMaterialNonce(ctx context.Context, chain string, nonce int) error {
	field := strings.ToLower(fmt.Sprintf("%s_%s", chain, "material"))
	if query := util.WithContextDb(ctx).Model(MemberTakeBack{}).Where(fmt.Sprintf("member_id = ? and %s = ?", field), m.ID, nonce).
		UpdateColumn(map[string]interface{}{field: gorm.Expr(fmt.Sprintf("%s + ?", field), 1)}); query == nil || query.RowsAffected == 0 {
		return errors.New("UpdateMaterialNonce fail")
	}
	return nil
}
