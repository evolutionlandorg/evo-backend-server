package models

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/jinzhu/gorm"
)

type Attribute struct {
	gorm.Model
	Title   string `json:"title"`
	TitleEn string `json:"title_en"`
	Kind    string `json:"kind"`
	Value   int    `json:"value"`
}

type AttributeJson struct {
	Id      uint   `json:"id"`
	Title   string `json:"title"`
	TitleEn string `json:"title_en"`
	Value   int    `json:"value"`
}

type ApostleAttribute struct {
	ApostleId   uint `json:"apostle_id"`
	AttributeId uint `json:"attribute_id"`
}

type SvgServerAttribute struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	NameEn string `json:"name_en"`
	Value  int    `json:"value"`
}

func (t *Attribute) New(ctx context.Context) error {
	result := util.WithContextDb(ctx).Create(&t)
	return result.Error
}

func (aa *ApostleAttribute) New(ctx context.Context) error {
	result := util.WithContextDb(ctx).Create(&aa)
	return result.Error
}

var SvgIdToKindMap = map[int]string{1: "profile", 2: "profile_color", 3: "feature", 4: "feature_color", 5: "hair", 6: "hair_color", 7: "eye", 8: "eye_color", 9: "expression", 10: "surroundings"}

func (ap *Apostle) createAttributeFromGene(ctx context.Context) {
	rb := util.HttpGet(fmt.Sprintf("%s/gn%s.svge", util.ApiServerHost, util.U256(ap.Genes).String()))
	var svgAttributes []SvgServerAttribute
	err := json.Unmarshal(rb, &svgAttributes)
	if err != nil || len(svgAttributes) < 1 {
		log.Debug("svgAttributes Unmarshal by json error: %s; rb: %s", err, string(rb))
		return
	}
	db := util.WithContextDb(ctx)
	for _, v := range svgAttributes {
		kind := SvgIdToKindMap[v.Id]
		if kind == "" {
			continue
		}
		var attribute Attribute
		query := db.First(&attribute, map[string]interface{}{"title": v.Name, "kind": kind})
		if query.RecordNotFound() {
			db.Create(&Attribute{Title: v.Name, TitleEn: v.NameEn, Kind: kind, Value: v.Value})
		} else {
			if attribute.TitleEn == "" || attribute.Value == 0 {
				db.Model(&attribute).UpdateColumn(map[string]interface{}{"title_en": v.NameEn, "value": v.Value})
			}
		}
		db.Create(ApostleAttribute{ApostleId: ap.ID, AttributeId: attribute.ID})
	}
}

func (ap *Apostle) getAttributes(ctx context.Context) map[string]AttributeJson {
	db := util.WithContextDb(ctx)
	var attributes []Attribute
	db.Model(&ap).Related(&attributes, "Attributes")
	aj := make(map[string]AttributeJson)
	for _, v := range attributes {
		aj[v.Kind] = AttributeJson{Title: v.Title, Id: v.ID, TitleEn: v.TitleEn, Value: v.Value}
	}
	return aj
}

func GetAttributeApostleId(ctx context.Context, attr int) []int {
	var ids []int
	db := util.WithContextDb(ctx)
	db.Table("apostle_attributes").Where("attribute_id =?", attr).Pluck("apostle_id", &ids)
	return ids
}
