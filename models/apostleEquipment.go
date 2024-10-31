package models

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
)

// Equipment 装备系统
type Equipment struct {
	ID               uint   `gorm:"primary_key"  json:"-" `
	EquipmentTokenId string `json:"equipment_token_id"`
	// Rarity 装备的品质, 木剑/钢剑/曙光女神之勇气
	Rarity int `json:"rarity"`
	// Level 装备的等级, 强化相关
	Level          int    `json:"level"`
	Prefer         string `json:"prefer"`
	Object         string `json:"object"`
	Owner          string `json:"owner"`
	ApostleTokenId string `json:"apostle_token_id"`
	Chain          string `json:"-"`
	Slot           int    `json:"slot"`         // 装备位置，默认1，目前只有1个位置
	OriginOwner    string `json:"origin_owner"` // 原主人，给使徒装备后owner会变
}

type EquipmentJson struct {
	EquipmentTokenId string         `json:"equipment_token_id"`
	Rarity           int            `json:"rarity"`
	Level            int            `json:"level"`
	Prefer           string         `json:"prefer"`
	Object           string         `json:"object"`
	Owner            string         `json:"owner"`
	OriginOwner      string         `json:"origin_owner"`
	Apostle          *ApostleSample `json:"apostle,omitempty"`
}

func GetEquipment(ctx context.Context, tokenId string) *Equipment {
	db := util.WithContextDb(ctx)
	var eq Equipment
	if query := db.Where("equipment_token_id  = ?", tokenId).Find(&eq); query.Error != nil || query.RecordNotFound() {
		return nil
	}
	return &eq
}

func apostle2Equipments(ctx context.Context, tokenId []string) map[string][]Equipment {
	if len(tokenId) < 1 {
		return nil
	}
	db := util.WithContextDb(ctx)
	var eqs []Equipment
	query := db.Where("apostle_token_id in (?)", tokenId).Find(&eqs)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	results := make(map[string][]Equipment)
	for _, v := range eqs {
		results[v.ApostleTokenId] = append(results[v.ApostleTokenId], v)
	}
	return results
}

func (e *Equipment) Transfer(ctx context.Context, dest, tokenId string) error {
	query := util.WithContextDb(ctx).Model(e).Where("equipment_token_id = ?", tokenId).
		UpdateColumn(Equipment{Owner: dest})
	return query.Error
}

func (e *Equipment) AsJson(ctx context.Context) *EquipmentJson {
	ej := EquipmentJson{
		EquipmentTokenId: e.EquipmentTokenId,
		Rarity:           e.Rarity,
		Level:            e.Level,
		Prefer:           e.Prefer,
		Object:           e.Object,
		Owner:            e.Owner,
		OriginOwner:      e.OriginOwner,
	}
	if e.ApostleTokenId != "" {
		if apostle := GetApostleByTokenId(ctx, e.ApostleTokenId); apostle != nil {
			ej.Apostle = &ApostleSample{ApostlePicture: apostle.ApostlePicture, Name: apostle.Name, Slot: e.Slot, ApostleTokenId: apostle.TokenId}
		}
	}
	return &ej
}

var objectMap = map[int]string{1: "Sword", 2: "Shield"}

func (ec *EthTransactionCallback) CraftBaseCallback(ctx context.Context) (err error) {

	if getTransactionDeal(ctx, ec.Tx, "CraftBase") != nil {
		return fmt.Errorf("tx exist")
	}
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("CraftBase", chain)) {
			logSlice := util.LogAnalysis(log.Data)
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			// Crafted(address to, uint256 tokenId, uint256 obj_id, uint256 rarity, uint256 prefer, uint256 timestamp)
			case services.AbiEncodingMethod("Crafted(address,uint256,uint256,uint256,uint256,uint256)"):

				objId := util.StringToInt(util.U256(logSlice[2]).String())
				rarity := util.StringToInt(util.U256(logSlice[3]).String())
				prefer := ""
				if preferValue := util.StringToInt(util.U256(logSlice[4]).String()); preferValue > 0 {
					prefer = preferMap[int(math.Log2(float64(preferValue)))-1]
				}
				db.Create(&Equipment{
					Owner:            util.AddHex(util.TrimHex(logSlice[0])[24:64], chain),
					EquipmentTokenId: logSlice[1],
					Rarity:           rarity,
					Prefer:           prefer,
					Object:           objectMap[objId],
					Level:            0,
					Chain:            chain,
				})
			// Enchanced(uint256 id, uint8 class, uint256 timestamp);
			case services.AbiEncodingMethod("Enchanced(uint256,uint8,uint256)"):
				db.Model(Equipment{}).Where("equipment_token_id = ?", logSlice[0]).
					Update("level", util.StringToInt(util.U256(logSlice[1]).String()))

			case services.AbiEncodingMethod("Disenchanted(uint256,uint8,uint256)"):
				db.Model(Equipment{}).Where("equipment_token_id = ?", logSlice[0]).
					Update("level", util.StringToInt(util.U256(logSlice[1]).String()))
			}
		}
	}
	if err != nil {
		return err
	}

	if err := ec.NewUniqueTransaction(db, "CraftBase"); err != nil {
		return err
	}

	db.DbCommit()
	return db.Error
}

func EquipmentList(ctx context.Context, opt ListOpt) (list []Equipment, count int) {
	query := util.WithContextDb(ctx).Model(Equipment{})
	for _, w := range opt.WhereQuery {
		query = query.Where(w)
	}
	query.Count(&count)
	order := util.TrueOrElse(opt.Order == "", "id desc", fmt.Sprintf("%s %s", opt.OrderField, opt.Order))
	query.Order(order).Offset(opt.Page * opt.Row).Limit(opt.Row).Find(&list)
	return
}

func SetEquipmentOriginOwner(ctx context.Context, tokenId, dest string) error {
	query := util.WithContextDb(ctx).Model(Equipment{}).Where("equipment_token_id = ?", tokenId).
		UpdateColumn(map[string]string{"origin_owner": dest})
	return query.Error
}

func SetEquipmentOriginOwnerByApostleId(db *util.GormDB, tokenId, dest string) error {
	query := db.Model(Equipment{}).Where("apostle_token_id = ?", tokenId).UpdateColumn(map[string]string{"origin_owner": dest})
	return query.Error
}
