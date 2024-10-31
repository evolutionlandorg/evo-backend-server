package models

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/shopspring/decimal"
)

type Drill struct {
	Id           uint       `gorm:"primary_key" json:"-"`
	CreateTime   int        `json:"create_time"`
	Owner        string     `json:"owner"`
	TokenId      string     `json:"token_id"`
	Class        int        `json:"class"`
	Grade        int        `json:"grade"`
	Prefer       string     `json:"prefer"`
	FormulaIndex int        `json:"formula_index"`
	FormulaId    int        `json:"formula_id"`
	OriginOwner  string     `json:"origin_owner"`
	LandEquip    *LandEquip `json:"land_equip" gorm:"-"`
	Chain        string     `json:"chain" sql:"default:'Eth'"`
}

type LandEquip struct {
	DrillTokenId string `json:"drill_token_id" gorm:"primary_key;auto_increment:false"`
	LandTokenId  string `json:"land_token_id"`
	Index        int    `json:"index"`
	Resource     string `json:"resource"`
	Owner        string `json:"owner"`
	OwnerName    string `json:"owner_name"`
	FormulaId    int    `json:"formula_id"`
	EquipTime    int64  `json:"equip_time"`
	Prefer       string `json:"prefer"`
}

var protectionPeriod = []int{0, 7, 14, 21}
var preferMap = []string{"gold", "wood", "water", "fire", "soil"}

func Formulas(ctx context.Context, chain string) []util.Formula {
	formulas := util.Evo.Formula[chain]
	sort.Slice(formulas[:], func(i, j int) bool {
		return formulas[i].Sort < formulas[j].Sort
	})
	drillStat := drillStat(ctx, chain)
	for index, formula := range formulas {
		formulas[index].Issued = drillStat[formula.Id]
		formulas[index].ProtectionPeriod = protectionPeriod[formula.Class]
	}
	return formulas

}

func getFormula(chain string, class, grade, ObjectClassExt int) util.Formula {
	for _, formula := range util.Evo.Formula[chain] {
		if formula.Class == class && formula.Grade == grade && formula.ObjectClassExt == ObjectClassExt {
			return formula
		}
	}
	return util.Formula{}
}

func CreateDrill(txn *util.GormDB, owner, tokenId string, createTime, class, grade, preferValue, FormulaIndex, extClass int, chain string) error {
	prefer := ""
	if preferValue > 0 {
		prefer = preferMap[int(math.Log2(float64(preferValue)))-1]
	}
	query := txn.Create(&Drill{
		Owner:        owner,
		TokenId:      tokenId,
		CreateTime:   createTime,
		Class:        class,
		Grade:        grade,
		Prefer:       prefer,
		FormulaIndex: FormulaIndex,
		FormulaId:    getFormula(chain, class, grade, extClass).Id,
		Chain:        chain,
	})
	if query.Error == nil {
		refreshOpenSeaMetadata(tokenId)
	}
	return query.Error
}

func Drills(ctx context.Context, opt ListOpt, owner string, multiFilter DrillMultiFilter) ([]Drill, int) {
	var (
		list []Drill
	)
	query := util.WithContextDb(ctx).Model(Drill{})

	for _, q := range opt.WhereQuery {
		query = query.Where(q)
	}
	if opt.Order != "" {
		query = query.Order(fmt.Sprintf("formula_id %s", opt.Order))
	}

	query.Find(&list)

	//count := len(list)
	// Pagination

	if opt.Display != "ignore" && opt.Chain == EthChain {
		// dego network
		sg := storage.New(EthChain)
		degoTokenId := sg.DEGOTokensOfOwner(owner)

		// filter items has working
		if opt.Filter != "fresh" {
			var equipDego []string
			util.WithContextDb(ctx).Model(LandEquip{}).Where("owner = ? and formula_id = ?", owner, 256).Pluck("drill_token_id", &equipDego)
			degoTokenId = append(degoTokenId, equipDego...)
		}

		for _, raw := range util.UniqueStrings(degoTokenId) {
			list = append(list, Drill{TokenId: raw, Class: 0, Grade: 1, Owner: owner, FormulaId: 256})
		}
	}
	list = filterDrills(list, multiFilter)
	count := len(list)
	if count == 0 {
		return nil, 0
	}
	maxLength := (opt.Page + 1) * opt.Row
	if maxLength > count {
		maxLength = count
	}
	list = list[opt.Page*opt.Row : maxLength]
	return list, count
}

type DrillMultiFilter struct {
	Class   []string
	Grade   []string
	Element []string
}

func filterDrills(list []Drill, multiFilter DrillMultiFilter) (drills []Drill) {
	for _, drill := range list {

		if len(multiFilter.Element) > 0 && len(multiFilter.Element) < 5 {
			var pass bool
			for _, element := range multiFilter.Element {
				if drill.Prefer == element {
					pass = true
					break
				}
			}
			if !pass {
				continue
			}
		}
		if len(multiFilter.Grade) > 0 && len(multiFilter.Grade) < 3 {
			var pass bool
			for _, grade := range multiFilter.Grade {
				if drill.Grade == util.StringToInt(grade) {
					pass = true
					break
				}
			}
			if !pass {
				continue
			}
		}
		if len(multiFilter.Class) > 0 && len(multiFilter.Class) < 3 {
			var pass bool
			for _, class := range multiFilter.Class {
				if drill.Class == util.StringToInt(class) {
					pass = true
					break
				}
			}
			if !pass {
				continue
			}
		}
		drills = append(drills, drill)
	}
	return
}

func GetDrillsByTokenId(ctx context.Context, tokenId string) *Drill {
	db := util.WithContextDb(ctx)
	var drill Drill
	query := db.Where("token_id  = ?", tokenId).Find(&drill)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	return &drill
}

func (d *Drill) Transfer(ctx context.Context, dest, tokenId string) error {
	if strings.EqualFold(util.GetContractAddress("landResource", d.Chain), d.Owner) &&
		!strings.EqualFold(util.GetContractAddress("landResource", d.Chain), dest) { // 抢占
		LandEquipRemove(ctx, d.TokenId)
	}
	query := util.WithContextDb(ctx).Model(d).Where("token_id = ?", tokenId).
		UpdateColumn(Drill{Owner: dest})
	return query.Error
}

func (d *Drill) RefreshFormulaId(ctx context.Context, extClass int) error {
	formula := getFormula(d.Chain, d.Class, d.Grade, extClass)
	if formula.Id == 0 {
		return fmt.Errorf("not found formula id chain=%s class=%d, grade=%d, extClass=%d",
			d.Chain, d.Class, d.Grade, extClass)
	}
	return util.WithContextDb(ctx).Model(d).Where("id = ?", d.Id).
		UpdateColumn(Drill{FormulaId: formula.Id}).Error
}

func (d *Drill) UnclaimedResource() map[string]decimal.Decimal {
	zero := decimal.Zero
	resources := map[string]decimal.Decimal{"gold": zero, "wood": zero, "water": zero, "fire": zero, "soil": zero}
	chain := GetChainByTokenId(d.TokenId)
	resourceAddress := []string{
		util.GetContractAddress("gold", chain),
		util.GetContractAddress("wood", chain),
		util.GetContractAddress("water", chain),
		util.GetContractAddress("fire", chain),
		util.GetContractAddress("soil", chain),
	}
	sg := storage.New(chain)
	result := sg.DrillUnclaimedResource(d.TokenId, resourceAddress)
	if len(result) >= 5 {
		resources["gold"] = util.BigToDecimal(util.U256(result[0]))
		resources["wood"] = util.BigToDecimal(util.U256(result[1]))
		resources["water"] = util.BigToDecimal(util.U256(result[2]))
		resources["fire"] = util.BigToDecimal(util.U256(result[3]))
		resources["soil"] = util.BigToDecimal(util.U256(result[4]))
	}
	return resources
}

func SetDrillOriginOwner(ctx context.Context, tokenId, dest string) error {
	query := util.WithContextDb(ctx).Model(Drill{}).Where("token_id = ?", tokenId).
		UpdateColumn(map[string]string{"origin_owner": dest})
	return query.Error
}

func drillStat(ctx context.Context, chain string) map[int]int {
	type GCount struct {
		Count     int
		FormulaId int
	}
	var countArr []GCount
	util.WithContextDb(ctx).Select("count(*) as count,formula_id").Model(Drill{}).Where("chain = ?", chain).Group("formula_id").Scan(&countArr)
	countMap := make(map[int]int)
	for _, v := range countArr {
		countMap[v.FormulaId] = v.Count
	}
	return countMap
}

func FullyLoadedLandId(ctx context.Context) []string {
	var tokenId []string
	util.WithContextDb(ctx).Model(LandEquip{}).Group("land_token_id").Having("COUNT(land_token_id)=3").Pluck("land_token_id", &tokenId)
	return tokenId
}

func drillProtectPeriod(ctx context.Context, tokenId string, class int, defaultEqTime int64) int64 {
	key := fmt.Sprintf("ProtectPeriod:%s", tokenId)
	now := time.Now().Unix()
	protection := protectionPeriod[class]
	if protection == 0 {
		return defaultEqTime
	}
	if cache := util.GetCache(ctx, key); len(cache) != 0 {
		return util.StringToInt64(string(cache))
	}
	chain := GetChainByTokenId(tokenId)
	st := storage.New(chain)
	if protectEnd := st.ProtectPeriod(util.GetContractAddress("objectOwnership", chain), tokenId); protectEnd == 0 {
		return defaultEqTime
	} else {
		eqTime := protectEnd - int64(protection)*86400
		_ = util.SetCache(ctx, key, []byte(util.IntToString(int(eqTime))), int(protectEnd-now))
		return eqTime
	}
}
