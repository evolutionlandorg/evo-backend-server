package models

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/evolutionlandorg/evo-backend/util/pve"

	"fmt"
	"math/big"
	"os"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/pkg/errors"

	"github.com/shopspring/decimal"
)

type NftMetaData struct {
	Description string                   `json:"description"`
	ExternalUrl string                   `json:"external_url"`
	Image       string                   `json:"image"`
	Name        string                   `json:"name"`
	Attributes  []map[string]interface{} `json:"attributes"`
}

func GetNftMetaData(ctx context.Context, tokenId string) (*NftMetaData, error) {
	var metaData NftMetaData
	if strings.HasPrefix(tokenId, "2a") || strings.HasPrefix(tokenId, "0x") {
		tokenId = util.U256(tokenId).String()
	}
	decimalTokenId, err := decimal.NewFromString(tokenId)
	if err != nil {
		return nil, errors.Wrap(err, "tokenId is not valid")
	}
	tokenId = util.DecodeInputU256(decimalTokenId.Coefficient())
	if len(tokenId) != 64 {
		return nil, errors.Wrap(err, "tokenId is not valid")
	}
	tokenTypeSymbol := getAssetTypeByTokenId(tokenId)
	switch tokenTypeSymbol {
	case AssetLand:

		land := GetLandByTokenId(ctx, tokenId)
		if land == nil {
			return nil, errors.New("no land found")
		}
		url := fmt.Sprintf("%s/land/%d/land/0x%s", os.Getenv("WEB_HOST"), land.District, land.TokenId)
		metaData = NftMetaData{
			Description: land.Introduction,
			ExternalUrl: url,
			Image:       GetLandPicture(land.District, land.GX, land.GY),
			Name:        land.landName(),
			Attributes:  land.metaDataAttributes(ctx),
		}
	case AssetApostle:
		apostle := GetApostleByTokenId(ctx, tokenId)
		if apostle == nil {
			return nil, errors.New("not found apostle")
		}
		url := fmt.Sprintf("%s/land/%d/apostle/0x%s", os.Getenv("WEB_HOST"), apostle.District, apostle.TokenId)
		metaData = NftMetaData{
			Description: apostle.Introduction,
			ExternalUrl: url,
			Name:        apostle.Name,
			Attributes:  apostle.metaDataAttributes(ctx),
		}
		if strings.HasPrefix(apostle.ApostlePicture, "/") {
			metaData.Image = GetApostlePicture(apostle.Genes, apostle.Chain, apostle.TokenIndex)
		}
	case AssetMirrorKitty:
		mirror := getPetTokenIdByMirror(ctx, tokenId)
		if mirror == nil {
			return nil, errors.New("not found mirror kitty")
		}
		metaData = NftMetaData{
			Name:       fmt.Sprintf("Mirror#%s", mirror.TokenId),
			Image:      "https://gcs.evolution.land/mirror/kittyMirror.png",
			Attributes: []map[string]interface{}{},
		}
	case AssetDrill, AssetItem:
		drill := GetDrillsByTokenId(ctx, tokenId)
		if drill == nil {
			return nil, errors.New("not found drill or item")
		}
		url := fmt.Sprintf("%s/land/%d/drill/0x%s", os.Getenv("WEB_HOST"), getNFTDistrict(drill.TokenId), drill.TokenId)
		for _, formula := range Formulas(ctx, drill.Chain) {
			if formula.Id == drill.FormulaId {
				metaData = NftMetaData{
					ExternalUrl: url,
					Name:        formula.Name,
					Description: "Powerful tool that could mine resources from every land",
					Image:       fmt.Sprintf("https://gcs.evolution.land/furnace/v2/%s", formula.Pic),
					Attributes:  drill.metaData(formula),
				}
			}
		}
	case Material:
		id := new(big.Int).Mod(util.U256(tokenId), big.NewInt(65536)).Int64()
		materialSymbol := pve.MaterialIdToSymbol(int(id))
		material := pve.GetStageConf().Material[materialSymbol]
		metaData = NftMetaData{
			// ExternalUrl: url,
			Name:        material.Name,
			Description: material.Desc,
			Image:       fmt.Sprintf("https://gcs.evolution.land/assets/material/%s.png", strings.ToUpper(materialSymbol)),
			Attributes:  []map[string]interface{}{{"rarity": material.Rarity}},
		}
	case AssetEquipment:
		eq := GetEquipment(ctx, tokenId)
		if eq == nil {
			return nil, fmt.Errorf("not found equipment token_id %s", tokenId)
		}
		equipments, ok := pve.GetStageConf(eq.Chain).Equipments[eq.Object]
		if !ok {
			return nil, errors.New("not found equipment from stage conf")
		}
		equipment, ok := equipments[util.IntToString(eq.Rarity)]
		if !ok {
			return nil, errors.New("not found equipment")
		}
		metaData = NftMetaData{
			ExternalUrl: fmt.Sprintf("%s/land/%d/equipment/0x%s", os.Getenv("WEB_HOST"), getNFTDistrict(eq.EquipmentTokenId), eq.EquipmentTokenId),
			Name:        equipment.Name,
			Description: equipment.Name,
			Image: fmt.Sprintf("https://gcs.evolution.land/assets/equipment/%s/lv%d/rarity%d.png",
				strings.ToLower(eq.Object), eq.Level, eq.Rarity),
			Attributes: eq.metaData(equipment),
		}
	default:
		return nil, errors.Errorf("get %s not found type %s", tokenId, tokenTypeSymbol)
	}
	return &metaData, nil
}

func (l *Land) metaDataAttributes(ctx context.Context) []map[string]interface{} {
	var attr []map[string]interface{}
	resource := landResources(ctx, l.TokenId)
	attr = append(attr, map[string]interface{}{"trait_type": "gold_volume", "display_type": "number", "value": resource.GoldRate})
	attr = append(attr, map[string]interface{}{"trait_type": "wood_volume", "display_type": "number", "value": resource.WoodRate})
	attr = append(attr, map[string]interface{}{"trait_type": "hho_volume", "display_type": "number", "value": resource.WaterRate})
	attr = append(attr, map[string]interface{}{"trait_type": "fire_volume", "display_type": "number", "value": resource.FireRate})
	attr = append(attr, map[string]interface{}{"trait_type": "soil_volume", "display_type": "number", "value": resource.SoilRate})
	if resource.IsReserved == 1 {
		attr = append(attr, map[string]interface{}{"trait_type": "is_reserved", "value": "reserve"})
	}
	return attr
}

func (ap *Apostle) metaDataAttributes(ctx context.Context) []map[string]interface{} {
	var attr []map[string]interface{}
	attr = append(attr, map[string]interface{}{"trait_type": "generation", "display_type": "number", "value": ap.Gen})
	attr = append(attr, map[string]interface{}{"trait_type": "cold_down", "display_type": "number", "value": ap.ColdDown})
	talents := ap.ApostleTalent(ctx)
	bt, _ := json.Marshal(talents)
	var jTalents map[string]decimal.Decimal
	_ = json.Unmarshal(bt, &jTalents)
	for name, value := range jTalents {
		attr = append(attr, map[string]interface{}{"trait_type": name, "display_type": "number", "value": value.InexactFloat64()})
	}
	apostleAttr := ap.getAttributes(ctx)
	for name, value := range apostleAttr {
		if value.TitleEn != "" {
			attr = append(attr, map[string]interface{}{"trait_type": name, "value": value.TitleEn})
		}
	}
	return attr
}

func (d *Drill) metaData(formula util.Formula) []map[string]interface{} {
	var attr []map[string]interface{}
	attr = append(attr, map[string]interface{}{"trait_type": "drill_class", "display_type": "number", "value": formula.Class})
	attr = append(attr, map[string]interface{}{"trait_type": "drill_grade", "display_type": "number", "value": formula.Grade})
	return attr
}

func (e *Equipment) metaData(equipment pve.Equipment) []map[string]interface{} {
	var attr []map[string]interface{}
	attr = append(attr, map[string]interface{}{"trait_type": "drill_class", "value": equipment.Class})
	attr = append(attr, map[string]interface{}{"trait_type": "rarity", "display_type": "number", "value": equipment.Rarity})
	return attr
}
