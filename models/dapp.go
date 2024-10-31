package models

import (
	"context"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
)

type Dapp struct {
	ID           uint `gorm:"primary_key"`
	CreatedAt    time.Time
	Status       string `json:"status"`
	MemberId     uint   `json:"member_id"`
	LandId       uint   `json:"land_id"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Introduction string `json:"introduction"`
	Cover        string `json:"cover"`
	Email        string `json:"email"`
	Url          string `json:"url"`
}

type DappJson struct {
	Name         string `json:"name"`
	Status       string `json:"status"`
	Category     string `json:"category"`
	Introduction string `json:"introduction"`
	Cover        string `json:"cover"`
	Email        string `json:"email"`
	Url          string `json:"url"`
}

type ValidateDapp struct {
	Name         string `form:"name" json:"name" binding:"required,max=30"`
	Category     string `form:"category" json:"category" binding:"required"`
	Introduction string `form:"introduction" json:"introduction" binding:"required,max=100"`
	Email        string `form:"email" json:"email" binding:"required,email"`
	Cover        string `form:"cover" json:"cover" binding:"required,url"`
	Url          string `form:"url" json:"url" binding:"required,url"`
}

type ValidateAddDapp struct {
	LandId uint `form:"land_id" json:"land_id" binding:"required"`
	ValidateDapp
}

type ValidateEditDapp struct {
	LandId uint `form:"land_id" json:"land_id" binding:"required"`
	ValidateDapp
	DappInstant *Dapp
}

type DappReport struct {
	gorm.Model
	DappId uint   `json:"dapp_id"`
	Reason string `json:"reason"`
	Remark string `json:"remark"`
}

var DappCategory = []string{"Game", "Social", "Tool", "MarketPlace", "Casino", "Other"}

const (
	DappSubmitted     = "submitted"
	DappSubmitSuccess = "success"
)

func (va *ValidateAddDapp) AddDapp(ctx context.Context, memberId uint, status ...string) error {
	db := util.WithContextDb(ctx)
	dapp := Dapp{MemberId: memberId, LandId: va.LandId, Name: va.Name, Category: va.Category, Introduction: va.Introduction, Cover: va.Cover, Email: va.Email, Url: va.Url}

	dapp.Status = DappSubmitted
	if len(status) != 0 && util.StringInSlice(status[0], []string{DappSubmitted, DappSubmitSuccess}) {
		dapp.Status = status[0]
	}
	query := db.Create(&dapp)
	return query.Error
}

func (va *ValidateEditDapp) EditDapp(ctx context.Context, status ...string) error {
	db := util.WithContextDb(ctx)
	dapp := Dapp{Name: va.Name, Category: va.Category, Introduction: va.Introduction, Cover: va.Cover, Email: va.Email, Url: va.Url}
	dapp.Status = DappSubmitted
	if len(status) != 0 && util.StringInSlice(status[0], []string{DappSubmitted, DappSubmitSuccess}) {
		dapp.Status = status[0]
	}
	dapp.MemberId = va.DappInstant.MemberId
	dapp.LandId = va.DappInstant.LandId
	query := db.Create(&dapp)
	return query.Error
}

func (dapp *Dapp) Del(txn *util.GormDB) error {
	query := txn.Where("land_id=?", dapp.LandId).Delete(Dapp{})
	return query.Error
}

func (dapp *Dapp) Report(ctx context.Context, reason []string, remark string) {
	db := util.WithContextDb(ctx)
	report := DappReport{Reason: strings.Join(reason, "|"), Remark: remark, DappId: dapp.ID}
	db.Create(&report)
}

func GetDappById(ctx context.Context, id uint) *Dapp {
	db := util.WithContextDb(ctx)
	var dapp Dapp
	query := db.First(&dapp, id)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	return &dapp
}

func GetDappByLandId(ctx context.Context, id uint) *Dapp {
	db := util.WithContextDb(ctx)
	var dapp Dapp
	query := db.First(&dapp, Dapp{LandId: id})
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	return &dapp
}
