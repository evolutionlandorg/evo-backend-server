package models

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
)

type LandApostle struct {
	ID         uint `gorm:"primary_key"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	LandId     uint            `json:"land_id"`
	ApostleId  uint            `json:"apostle_id"`
	DigElement string          `json:"dig_element"`
	Strength   decimal.Decimal `json:"strength" sql:"type:decimal(36,18);"`
}
type LandApostleJson struct {
	LandId     uint            `json:"land_id"`
	ApostleId  uint            `json:"apostle_id"`
	DigElement string          `json:"dig_element"`
	Lon        int             `json:"lon"`
	Lat        int             `json:"lat"`
	Strength   decimal.Decimal `json:"strength" sql:"type:decimal(36,18);"`
}

type ApostleMiner struct {
	ApostleId      uint               `json:"apostle_id"` // models.Apostle.ID
	TokenId        string             `json:"token_id"`   // models.Apostle.TokenId
	TokenIndex     int                `json:"token_index"`
	ApostleName    string             `json:"apostle_name"`
	OwnerName      string             `json:"owner_name"`
	Owner          string             `json:"owner"`
	ApostlePicture string             `json:"apostle_picture"`
	DigElement     string             `json:"dig_element"`
	Gen            int                `json:"gen"`
	ColdDown       int                `json:"cold_down"`
	Strength       decimal.Decimal    `json:"strength" sql:"type:decimal(36,18);"`
	Pets           *ApostlePetJson    `json:"pets"`
	Talent         *ApostleTalentJson `json:"talent"`
}

func CreateLandEquip(ctx context.Context, txn *util.GormDB, chain string, landTokenId string, index int, address, drillTokenId, resource string, blockTime int64) error {
	eq := &LandEquip{
		DrillTokenId: drillTokenId,
		LandTokenId:  landTokenId,
		Index:        index,
		Resource:     checkResourceType(resource, GetChainByTokenId(landTokenId)),
		Owner:        address,
		EquipTime:    blockTime,
		FormulaId:    256,
	}
	if drill := GetDrillsByTokenId(ctx, drillTokenId); drill != nil {
		eq.FormulaId = drill.FormulaId
		eq.Prefer = drill.Prefer
	}
	if member := GetMemberByAddress(ctx, eq.Owner, chain); member != nil {
		eq.OwnerName = member.Name
	}
	query := txn.Create(eq)
	if query.RowsAffected > 0 {
		_ = SetDrillOriginOwner(ctx, drillTokenId, address)
	}
	if err := query.Error; err != nil {
		mysqlErr, ok := err.(*mysqlDriver.MySQLError)
		if ok && mysqlErr.Number == 1062 {
			return nil
		}
		return err
	}
	return nil
}

func LandEquipRemove(ctx context.Context, drillTokenId string) {
	if query := util.WithContextDb(ctx).Delete(LandEquip{}, "drill_token_id = ?", drillTokenId); query.RowsAffected > 0 {
		_ = SetDrillOriginOwner(ctx, drillTokenId, "")
	}
}

func LandEquipByTokenId(ctx context.Context, DrillTokenId []string) map[string]LandEquip {
	if len(DrillTokenId) == 0 {
		return nil
	}
	var eqs []LandEquip
	util.WithContextDb(ctx).Model(LandEquip{}).Where("drill_token_id in (?)", DrillTokenId).Find(&eqs)
	equipToLandId := make(map[string]LandEquip)
	for _, eq := range eqs {
		equipToLandId[eq.DrillTokenId] = eq
	}
	return equipToLandId
}

func (la *LandApostle) new(ctx context.Context) error {
	result := util.WithContextDb(ctx).Create(&la)
	return result.Error
}

func (ec *EthTransactionCallback) LandResourceCallback(ctx context.Context) (err error) {
	if getTransactionDeal(ctx, ec.Tx, "landResource") != nil {
		return errors.New("tx exist")
	}
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("landResource", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			// event StartMining(uint256 minerTokenId, uint256 landTokenId, address _resource, uint256 strength);
			case services.AbiEncodingMethod("StartMining(uint256,uint256,address,uint256)"):
				logSlice := util.LogAnalysis(log.Data)
				tokenId := logSlice[0]
				apostle := GetApostleByTokenId(ctx, tokenId)
				db.Model(&apostle).UpdateColumn("status", apostleWorking)
				land := getLand(ctx, logSlice[1])
				currencyMap := util.GetContractsMap(chain)
				lp := LandApostle{LandId: land.ID, ApostleId: apostle.ID, DigElement: currencyMap[strings.ToLower(util.AddHex(logSlice[2][24:64], chain))], Strength: util.BigToDecimal(util.U256(logSlice[3]))}
				_ = lp.new(ctx)
				apostle.startHabergerMode(db)
			case services.AbiEncodingMethod("StopMining(uint256,uint256,address,uint256)"):
				logSlice := util.LogAnalysis(log.Data)
				tokenId := logSlice[0]
				apostle := GetApostleByTokenId(ctx, tokenId)
				status := apostleFresh
				if auction := GetApostleWork(ctx, tokenId, AuctionFinish); auction != nil { // 在租赁中，状态改为Hiring
					status = apostleHiring
				}
				db.Model(&apostle).Where("status = ?", apostleWorking).UpdateColumn("status", status)
				land := getLand(ctx, logSlice[1])
				currencyMap := util.GetContractsMap(chain)
				lp := LandApostle{LandId: land.ID, ApostleId: apostle.ID, DigElement: currencyMap[strings.ToLower(util.AddHex(logSlice[2][24:64], chain))]}
				db.Table("land_apostles").Where(lp).Delete(&LandApostle{})
			case services.AbiEncodingMethod("ResourceClaimed(address,uint256,uint256,uint256,uint256,uint256,uint256)"):
				logSlice := util.LogAnalysis(log.Data)
				address := util.AddHex(util.TrimHex(logSlice[0])[24:64], chain)
				tokenId := logSlice[1]
				land := GetLandByTokenId(ctx, tokenId)
				goldBalance := util.BigToDecimal(util.U256(logSlice[2]))
				th := TransactionHistory{Tx: ec.Tx, Chain: chain, BalanceAddress: address, BalanceChange: goldBalance, Currency: currencyGold, Action: TransactionHistoryClaimResource, TokenId: tokenId, Coordinate: fmt.Sprintf("%d,%d", land.Lon, land.Lat)}
				if goldBalance.Sign() > 0 {
					_ = th.New(db)
				}
				woodBalance := util.BigToDecimal(util.U256(logSlice[3]))
				if woodBalance.Sign() > 0 {
					th.BalanceChange = woodBalance
					th.Currency = currencyWood
					_ = th.New(db)
				}
				waterBalance := util.BigToDecimal(util.U256(logSlice[4]))
				if waterBalance.Sign() > 0 {
					th.BalanceChange = waterBalance
					th.Currency = currencyWater
					_ = th.New(db)
				}
				fireBalance := util.BigToDecimal(util.U256(logSlice[5]))
				if fireBalance.Sign() > 0 {
					th.BalanceChange = fireBalance
					th.Currency = currencyFire
					_ = th.New(db)
				}
				soilBalance := util.BigToDecimal(util.U256(logSlice[6]))
				if soilBalance.Sign() > 0 {
					th.BalanceChange = soilBalance
					th.Currency = currencySoil
					_ = th.New(db)
				}
			case services.AbiEncodingMethod("UpdateMiningStrengthWhenStart(uint256,uint256,uint256)"):
				logSlice := util.LogAnalysis(log.Data)
				tokenId := logSlice[0]
				land := getLand(ctx, logSlice[1])
				apostle := GetApostleByTokenId(ctx, tokenId)
				lp := LandApostle{LandId: land.ID, ApostleId: apostle.ID}
				db.Table("land_apostles").Where(lp).UpdateColumn(LandApostle{Strength: util.BigToDecimal(util.U256(logSlice[2]))})
				// index tokenId, resource, index, staker, token, id
			case services.AbiEncodingMethod("Equip(uint256,address,uint256,address,address,uint256)"):
				logSlice := util.LogAnalysis(log.Data)
				landTokenId := util.TrimHex(log.Topics[1])
				index := util.StringToInt(util.U256(logSlice[1]).String())
				address := util.AddHex(logSlice[2][24:64], chain)
				drillTokenId := logSlice[4]
				resource := util.AddHex(logSlice[0][24:64], chain)
				db.Delete(LandEquip{}, "land_token_id = ? and `index` = ?", landTokenId, index)
				err = CreateLandEquip(ctx, db, chain, landTokenId, index, address, drillTokenId, resource, ec.BlockTimestamp)
			case services.AbiEncodingMethod("Divest(uint256,address,uint256,address,address,uint256)"):
				logSlice := util.LogAnalysis(log.Data)
				LandEquipRemove(ctx, logSlice[4])
			}
		}
	}
	if err != nil {
		return
	}

	if err := ec.NewUniqueTransaction(db, "landResource"); err != nil {
		return err
	}

	db.DbCommit()
	return db.Error
}

func (l *LandJson) apostles(ctx context.Context) []ApostleMiner {
	db := util.WithContextDb(ctx)
	var miners []ApostleMiner
	var apostleIds []uint
	db.Table("land_apostles").Where("land_id =?", l.ID).Scan(&miners)
	for _, v := range miners {
		apostleIds = append(apostleIds, v.ApostleId)
	}
	apostles := getApostleReversalKeyByIds(ctx, apostleIds)
	bindPetApostle := getApostlePetsReversalKey(ctx, apostleIds)
	apostleTalents := getApostleTalentReversalKey(ctx, apostleIds)
	chain := GetChainByTokenId(l.TokenId)
	for k, v := range miners {
		miners[k].TokenId = apostles[v.ApostleId].TokenId
		miners[k].ApostlePicture = apostles[v.ApostleId].ApostlePicture
		miners[k].ApostlePicture = GetApostlePicture(apostles[v.ApostleId].Genes, apostles[v.ApostleId].Chain, apostles[v.ApostleId].TokenIndex)
		miners[k].TokenIndex = apostles[v.ApostleId].TokenIndex
		miners[k].ApostleName = apostles[v.ApostleId].Name
		miners[k].Pets = nil
		miners[k].Gen = apostles[v.ApostleId].Gen
		miners[k].ColdDown = apostles[v.ApostleId].ColdDown
		if talent, ok := apostleTalents[v.ApostleId]; ok {
			miners[k].Talent = &talent.ApostleTalentJson
		}

		if bindPetApostle != nil && bindPetApostle[v.ApostleId].ApostleId != 0 {
			pet := bindPetApostle[v.ApostleId]
			miners[k].Pets = &pet
		}
		miners[k].Owner = apostles[v.ApostleId].Owner
		if member := GetMemberByAddress(ctx, apostles[v.ApostleId].Owner, chain); member != nil {
			miners[k].OwnerName = member.Name
		}
	}
	return miners
}

func (ap *Apostle) ApostleLandDig(ctx context.Context) *LandApostleJson {
	db := util.WithContextDb(ctx)
	if ap.Status != apostleWorking {
		return nil
	}
	var lj LandApostleJson
	query := db.Table("land_apostles").Where("apostle_id =?", ap.ID).Scan(&lj)
	if query.RecordNotFound() {
		return nil
	}
	var land Land
	db.Model(&Land{}).Where("id =?", lj.LandId).Scan(&land)
	lj.Lon = land.Lon
	lj.Lat = land.Lat
	return &lj
}

func checkResourceType(resourceAddress, chain string) string {
	switch resourceAddress {
	case util.GetContractAddress("gold", chain):
		return "gold"
	case util.GetContractAddress("wood", chain):
		return "wood"
	case util.GetContractAddress("water", chain):
		return "water"
	case util.GetContractAddress("fire", chain):
		return "fire"
	case util.GetContractAddress("soil", chain):
		return "soil"
	}
	return ""
}

func DigLands(ctx context.Context) (landIds []string) {
	util.WithContextDb(ctx).Model(LandApostle{}).Pluck("land_id", &landIds)
	return
}

func (m *Member) EqDrill(ctx context.Context, wallet string) (tokenIds []string) {
	util.WithContextDb(ctx).Model(LandEquip{}).Where("owner = ?", wallet).Pluck("drill_token_id", &tokenIds)
	return
}
