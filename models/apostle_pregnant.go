package models

import (
	"context"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
)

type ApostlePregnant struct {
	gorm.Model
	MatronTokenId string `json:"matron_token_id"`
	SireTokenId   string `json:"sire_token_id"`
	BabyTokenId   string `json:"baby_token_id"`
	Tx            string `json:"tx"`
}

func (ap *ApostlePregnant) New(db *util.GormDB) error {
	result := db.Create(&ap)
	return result.Error
}

func (ap *Apostle) updateCoolDownEnd(db *util.GormDB, coolDown, coolDownEnd int) {
	db.Model(&ap).UpdateColumn(Apostle{ColdDown: coolDown, ColdDownEnd: coolDownEnd, HabergerMode: 1})
}

func dealApostlePregnant(ctx context.Context, db *util.GormDB, tx string, matronTokenId, sireTokenId string, matronCoolDown, sireCoolDown, matronCoolDownEndTime, sireCoolDownEndTime int) error {
	aPregnant := ApostlePregnant{MatronTokenId: matronTokenId, SireTokenId: sireTokenId, Tx: tx}
	_ = aPregnant.New(db)
	mother := GetApostleByTokenId(ctx, matronTokenId)
	mother.updateCoolDownEnd(db, matronCoolDown, matronCoolDownEndTime)
	apostle := Apostle{Owner: mother.Owner, MemberId: mother.MemberId, Duration: matronCoolDownEndTime, Status: apostleBirth, Mother: matronTokenId, Father: sireTokenId, District: getNFTDistrict(matronTokenId)}
	father := GetApostleByTokenId(ctx, sireTokenId)
	father.updateCoolDownEnd(db, sireCoolDown, sireCoolDownEndTime)
	th := TransactionHistory{Tx: tx, Chain: GetChainByTokenId(matronTokenId), BalanceAddress: mother.Owner, BalanceChange: apostleBirthFee.Neg(), Action: TransactionHistoryApostlePregnant, TokenId: matronTokenId, Currency: currencyRing}
	_ = th.New(db)
	// 怀孕后自动增加一条不含token_id的Apostle
	return apostle.New(db)
}
