package models

import (
	"context"
	"github.com/evolutionlandorg/evo-backend/util/pve"
	"math/big"

	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type ApostleTalent struct {
	gorm.Model `json:"-"`
	TokenId    string `json:"token_id"`
	ApostleId  uint   `json:"apostle_id"`
	ApostleTalentJson
}

type ApostleTalentJson struct {
	Life         *int   `json:"life"`
	LifeAdd      *int   `json:"life_add"`
	Mood         *int   `json:"mood"`
	MoodAdd      *int   `json:"mood_add"`
	Strength     *int   `json:"strength"`
	StrengthAdd  *int   `json:"strength_add"`
	Agile        *int   `json:"agile"`
	AgileAdd     *int   `json:"agile_add"`
	Finesse      *int   `json:"finesse"`
	FinesseAdd   *int   `json:"finesse_add"`
	Hp           *int   `json:"hp"`
	HpAdd        *int   `json:"hp_add"`
	Intellect    *int   `json:"intellect"`
	IntellectAdd *int   `json:"intellect_add"`
	Lucky        *int   `json:"lucky"`
	LuckyAdd     *int   `json:"lucky_add"`
	Potential    *int   `json:"potential"`
	PotentialAdd *int   `json:"potential_add"`
	ElementGold  *int   `json:"element_gold"`
	ElementWood  *int   `json:"element_wood"`
	ElementWater *int   `json:"element_water"`
	ElementFire  *int   `json:"element_fire"`
	ElementSoil  *int   `json:"element_soil"`
	Skills       *int   `json:"skills"`
	Secret       *int   `json:"secret"`
	Charm        *int   `json:"charm"`
	CharmAdd     *int   `json:"charm_add"`
	Expansion    *int64 `json:"expansion"`
	// 挖矿力
	MiningPower decimal.Decimal `json:"mining_power" sql:"type:decimal(15,4);"`
	// pve value
	ATK     decimal.Decimal `json:"atk" sql:"type:decimal(15,6);"`
	CRIT    decimal.Decimal `json:"crit" sql:"type:decimal(15,6);"`
	HPLimit decimal.Decimal `json:"hp_limit" sql:"type:decimal(15,6);"`
	DEF     decimal.Decimal `json:"def" sql:"type:decimal(15,6);"`
}

func (at *ApostleTalent) New(db *util.GormDB) error {
	result := db.Create(&at)
	return result.Error
}

func (ap *Apostle) createApostleTalentFromTalent(db *util.GormDB, talent string) error {
	at := ApostleTalent{TokenId: ap.TokenId}
	at.ApostleTalentJson = *ApostleTalentObject(talent)
	at.ApostleId = ap.ID
	err := at.New(db)
	return err
}

func (ap *Apostle) RefreshTalent(txn *util.GormDB, talent string) error {
	at := ApostleTalentObject(talent)
	var talentInstant ApostleTalent
	query := txn.Model(&talentInstant).Where("token_id =?", ap.TokenId).Updates(ApostleTalent{ApostleTalentJson: *at})
	return query.Error
}

func ApostleTalentObject(talent string) *ApostleTalentJson {
	talents := talentDecode(talent)
	at := ApostleTalentJson{}
	at.Strength = &talents[0]
	at.Agile = &talents[1]
	at.Intellect = &talents[2]
	at.Life = &talents[3]
	at.Hp = &talents[4]
	at.Mood = &talents[5]
	at.Finesse = &talents[6]
	at.Lucky = &talents[7]
	at.Potential = &talents[8]
	at.ElementGold = &talents[9]
	at.ElementWood = &talents[10]
	at.ElementWater = &talents[11]
	at.ElementFire = &talents[12]
	at.ElementSoil = &talents[13]
	at.Charm = &talents[17]
	at.StrengthAdd = &talents[18]
	at.AgileAdd = &talents[19]
	at.IntellectAdd = &talents[20]
	at.LifeAdd = &talents[21]
	at.HpAdd = &talents[22]
	at.MoodAdd = &talents[23]
	at.FinesseAdd = &talents[24]
	at.LuckyAdd = &talents[25]
	at.PotentialAdd = &talents[26]
	at.CharmAdd = &talents[29]

	at.ATK = pve.CalcATK(*at.Strength, *at.Intellect, *at.Finesse)
	at.CRIT = pve.CalcCRIT(*at.Lucky, *at.Mood)
	at.DEF = pve.CalcDEF(*at.Life, *at.Agile)
	at.HPLimit = pve.CalcHPLimit(*at.Hp, *at.Charm)

	at.MiningPower = GetApostleMiningPower(
		decimal.NewFromInt(int64(*at.Strength)),
		decimal.NewFromInt(int64(*at.Agile)),
		decimal.NewFromInt(int64(*at.Potential)),
	)
	return &at
}

func talentDecode(r string) []int {
	talentBig := util.U256(r)
	var talents []int
	for i := 0; i < 9; i++ {
		fi := new(big.Int).Lsh(big.NewInt(255), uint(8*i))
		si := new(big.Int).And(talentBig, fi)
		ci := new(big.Int).Rsh(si, uint(8*i))
		talents = append(talents, int(ci.Int64()))
	}
	prefer := new(big.Int).Rsh(talentBig, 72)
	prefer = new(big.Int).And(prefer, big.NewInt(65535))
	for i := 0; i < 8; i++ {
		fi := new(big.Int).Lsh(big.NewInt(3), uint(2*i))
		si := new(big.Int).And(prefer, fi)
		ci := new(big.Int).Rsh(si, uint(2*i))
		talents = append(talents, int(ci.Int64()))
	}
	charm := new(big.Int).Rsh(talentBig, 88)
	charm = new(big.Int).And(charm, big.NewInt(255))
	talents = append(talents, int(charm.Int64()))

	for i := 16; i < 28; i++ {
		fi := new(big.Int).Rsh(talentBig, uint(8*i))
		ci := new(big.Int).And(fi, big.NewInt(255))
		talents = append(talents, int(ci.Int64()))
	}
	return talents
}

func getApostleTalentReversalKey(ctx context.Context, ids []uint) map[uint]ApostleTalent {
	if len(ids) < 1 {
		return nil
	}
	db := util.WithContextDb(ctx)
	var pj []ApostleTalent
	query := db.Model(ApostleTalent{}).Where("apostle_id in (?)", ids).Scan(&pj)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	results := make(map[uint]ApostleTalent)
	for _, v := range pj {
		results[v.ApostleId] = v
	}
	return results
}
