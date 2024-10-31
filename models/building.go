package models

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"
)

type Building struct {
	ID                   uint   `gorm:"primary_key"  json:"id" `
	TokenId              string `json:"token_id"`
	Name                 string `json:"name"`
	Level                int    `json:"level"`
	DrawingIndex         int    `json:"drawing_index"`
	Address              string `json:"address"`
	Status               string `json:"status"` // idle/constructing/upgrading
	HiringStatus         bool   `json:"hiring_status"`
	AdminHiringStatus    bool   `json:"admin_hiring_status"`
	ConstructionStartAt  int    `json:"construction_start_at"`
	ConstructionDuration int    `json:"construction_duration"`
	ConstructionSub      int    `json:"construction_sub"`
	LandTokenId          string `json:"land_token_id"`
	LandLon              int    `json:"land_lon"`
	LandLat              int    `json:"land_lat"`
	Style                int    `json:"style"`
	AdminLevel           int    `json:"admin_level"`
	WorkerNum            uint   `json:"worker_num"`
	Location             uint   `json:"location"` // todo
}

type Drawing struct {
	Index                int             `json:"index"`
	Ring                 decimal.Decimal `json:"ring"`
	Kton                 decimal.Decimal `json:"kton"`
	Gold                 decimal.Decimal `json:"gold"`
	Wood                 decimal.Decimal `json:"wood"`
	Hoo                  decimal.Decimal `json:"hoo"`
	Fire                 decimal.Decimal `json:"fire"`
	Soil                 decimal.Decimal `json:"soil"`
	UpgradeDuration      []int           `json:"upgrade_duration"`
	LandRequire          int             `json:"land_require"`
	AdminLimit           int             `json:"admin_limit"`
	AdminEmploymentLimit int             `json:"admin_employment_limit"`
	AdminBuff            int             `json:"admin_buff"`
	LocationBuff         []int           `json:"location_buff"`
	MinerLimit           int             `json:"miner_limit"`
	Talent               []string        `json:"talent"`
	AdminUpgrade         []int           `json:"admin_upgrade"`
}

type BuildingHiring struct {
	BuildingId    uint            `json:"building_id" gorm:"primary_key"`
	Payment       string          `json:"payment"`
	PaymentAmount decimal.Decimal `json:"payment_amount" sql:"type:decimal(32,16);"`
	TalentLimit   uint            `json:"talent_limit"`
	WorkerLimit   uint            `json:"worker_limit"`
}

type BuildingAdminHiring struct {
	BuildingId    uint            `json:"building_id" gorm:"primary_key"`
	Payment       string          `json:"payment"`
	PaymentAmount decimal.Decimal `json:"payment_amount" sql:"type:decimal(32,16);"`
	TalentLimit   uint            `json:"talent_limit"`
	AdminLimit    uint            `json:"admin_limit"`
	Open          bool            `json:"open"`
	Duration      uint            `json:"duration"`
	AdminNum      uint            `json:"admin_num"`
	EmploymentNum uint            `json:"employment_num"`
}

type BuildingAdmin struct {
	BuildingId     uint   `json:"building_id"`
	ApostleTokenId string `json:"apostle_token_id"`
	IsEmployment   bool   `json:"is_employment"`
}

type BuildingWorker struct {
	BuildingId     uint   `json:"building_id"`
	ApostleTokenId string `json:"apostle_token_id"`
}

type BuildingDetail struct {
	*Building
	*Element
	AdminBuff    int                      `json:"admin_buff"`
	MiningBuff   int                      `json:"mining_buff"`
	OwnerName    string                   `json:"owner_name"`
	Talent       map[string]int           `json:"talent"`
	LocationBuff int                      `json:"location_buff"`
	Hiring       *BuildingHiring          `json:"hiring"`
	AdminHiring  *BuildingAdminHiring     `json:"admin_hiring"`
	Workers      []BuildingApostle        `json:"workers"`
	Admins       []BuildingApostle        `json:"admins"`
	Miner        []ApostleMiner           `json:"miner"`
	Dapp         *DappJson                `json:"dapp"`
	Auction      *AuctionJson             `json:"auction"`
	Record       []LandAuctionHistoryJson `json:"record"`
}

type BuildingApostle struct {
	TokenId        string `json:"token_id"`
	ApostlePicture string `json:"apostle_picture"`
	IsEmployment   bool   `json:"is_employment"`
	TalentPoint    int    `json:"talent_point"`
	TokenIndex     int    `json:"token_index"`
}

type Element struct {
	GoldRate  int `json:"gold_rate"`
	WoodRate  int `json:"wood_rate"`
	WaterRate int `json:"water_rate"`
	FireRate  int `json:"fire_rate"`
	SoilRate  int `json:"soil_rate"`
}

var drawings []Drawing

const (
	Idle             = "idle"
	Constructing     = "constructing"
	Upgrading        = "upgrading"
	perWorkerSubTime = 60 * 60
)

func GetDrawings() []Drawing {
	if len(drawings) > 0 {
		return drawings
	}
	b, err := os.ReadFile("data/buildingDrawing.json")
	util.Panic(err)
	util.UnmarshalAny(&drawings, b)
	return drawings
}

// create building level 0
func (d *Drawing) CreateBuilding(ctx context.Context, tokenId string, wallet string) error {
	land := getLand(ctx, tokenId)
	if land == nil || land.BuildingId > 0 {
		return errors.New("error land tokenId or this land already building")
	}
	coords := checkLandNear(land.Lon, land.Lat, d.Index)

	xRange := []int{-112, 68}
	yRange := []int{-22, 22}
	if land.District == 2 {
		xRange = []int{68, 112}
		yRange = []int{-22, 22}
	}
	var tokenIds []string
	for _, coord := range coords {
		coordSlice := strings.Split(coord, ",")
		x := util.StringToInt(coordSlice[0])
		y := util.StringToInt(coordSlice[1])
		if x < xRange[0] || x > xRange[1] || y < yRange[0] || y > yRange[1] {
			log.Debug("CreateBuilding info. x=%d. y=%d. xRange=%v; yRange=%v", x, y, xRange, yRange)
			return errors.New("land choose out of range")
		}
		land := GetLandByCoord(ctx, x, y)
		if !strings.EqualFold(land.Owner, wallet) || land.Status != landFresh || land.BuildingId > 0 {
			return errors.New("error land tokenId or this land already building")
		}
		tokenIds = append(tokenIds, land.TokenId)
	}
	txn := util.DbBegin(ctx)
	defer txn.Rollback()
	building := Building{
		DrawingIndex:         d.Index,
		Address:              wallet,
		Status:               Constructing,
		ConstructionStartAt:  int(time.Now().Unix()),
		ConstructionDuration: d.UpgradeDuration[0],
		LandTokenId:          tokenId,
		LandLon:              land.Lon,
		LandLat:              land.Lat,
		TokenId:              interstellarEncoding(BuildingContractId, land.District, len(Builds(ctx, nil))),
	}
	if tx := txn.Create(&building); tx.Error != nil {
		return tx.Error
	}
	if err := setLandBuilding(txn, building.ID, tokenIds); err != nil {
		return err
	}
	txn.DbCommit()
	return nil
}

func Builds(ctx context.Context, opt *ListOpt) []Building {
	var list []Building
	query := util.WithContextDb(ctx).Model(Building{})
	if opt != nil {
		if opt.OrderField != "" && opt.Order != "" {
			query = query.Order(fmt.Sprintf("%s %s", opt.OrderField, opt.Order))
		}
		if opt.Row > 0 {
			query = query.Offset(opt.Page * opt.Row).Limit(opt.Row)
		}
		for _, q := range opt.WhereQuery {
			query = query.Where(q)
		}
	}
	query.Find(&list)
	return list
}

//func GetBuilding(buildingId uint) *Building {
//	var building Building
//	if query := util.DB.First(&building, buildingId); query.RecordNotFound() {
//		return nil
//	}
//	return &building
//}

func (b *Building) AsJson(ctx context.Context, _ *Member) *BuildingDetail {
	var detail BuildingDetail
	detail.Building = b

	if util.StringInSlice(b.Status, []string{Constructing, Upgrading}) {
		detail.Hiring = b.Hiring(ctx)
		if detail.Hiring != nil {
			workers := detail.Hiring.Workers(ctx)
			detail.Workers = b.renderBuildingApostle(ctx, workers)
		}
	}
	if b.Status == Idle && b.AdminHiringStatus {
		detail.AdminHiring = b.AdminHiring(ctx)
		if detail.AdminHiring != nil {
			admins := detail.AdminHiring.Admins(ctx)
			detail.Admins = b.renderBuildingApostle(ctx, admins)
		}
	}

	if member := GetMemberByAddress(ctx, detail.Address, GetChainByTokenId(b.TokenId)); member != nil {
		detail.OwnerName = member.Name
	}

	// Todo hard code
	detail.Style = 1
	detail.AdminLevel = len(detail.Admins) % 3
	detail.AdminBuff = detail.AdminLevel * 5
	resource := landResources(ctx, detail.LandTokenId)
	var ele Element
	ele.GoldRate = resource.GoldRate
	ele.WoodRate = resource.WoodRate
	ele.WaterRate = resource.WaterRate
	ele.FireRate = resource.FireRate
	ele.SoilRate = resource.SoilRate
	detail.Element = &ele
	drawing := b.Drawing()
	talent := make(map[string]int, 3)
	for _, t := range drawing.Talent {
		talent[t] = 5
	}
	detail.Talent = talent
	detail.Location = 0 // todo
	detail.LocationBuff = drawing.LocationBuff[detail.Location]
	detail.MiningBuff = 100

	// if land := b.land(); land != nil {
	// 	detail.Dapp = land.Dapp(member)
	// 	auction := GetCurrentAuction(land.TokenId, AuctionGoing)
	// 	if auction != nil {
	// 		detail.Auction = auction.AsJson()
	// 	}
	// 	lj := LandJson{LandId: int64(land.TokenIndex + 1)}
	// 	detail.Miner = lj.apostles(nil)
	// 	detail.Record = landAuctionHistory(land.TokenId)
	// }
	return &detail
}

//func (b *Building) land() *Land {
//	return getLand(b.LandTokenId)
//}

func (b *Building) Upgrade(ctx context.Context) {
	util.WithContextDb(ctx).Model(b).Update(Building{
		Status:               Upgrading,
		ConstructionStartAt:  int(time.Now().Unix()),
		ConstructionDuration: b.Drawing().UpgradeDuration[b.Level],
	})
}

func (b *Building) Drawing() *Drawing {
	return &GetDrawings()[b.DrawingIndex-1]
}

func (b *Building) UpgradeComplete(ctx context.Context) {
	if query := util.WithContextDb(ctx).Model(b).Where("level = ?", b.Level).Update(map[string]interface{}{
		"status":        Idle,
		"level":         b.Level + 1,
		"hiring_status": false,
	}); query.RowsAffected > 0 {
		_ = b.freeWorker(ctx)
	}
}

func (b *Building) CancelUpgrade(ctx context.Context) {
	if query := util.WithContextDb(ctx).Model(b).Update(map[string]interface{}{
		"status":        Idle,
		"hiring_status": false,
	}); query.RowsAffected > 0 {
		_ = b.freeWorker(ctx)
	}
}

func (b *Building) setHiring(ctx context.Context) {
	util.WithContextDb(ctx).Model(b).Update(Building{HiringStatus: true})
}

func (b *Building) setAdminHiring(ctx context.Context) {
	util.WithContextDb(ctx).Model(b).Update(Building{AdminHiringStatus: true})
}

func (b *Building) SetName(ctx context.Context, name string) {
	util.WithContextDb(ctx).Model(b).Update(Building{Name: name})
}

func (b *Building) CreateHiring(ctx context.Context, payment string, amount decimal.Decimal, talentLimit, workerLimit uint) error {
	tx := util.WithContextDb(ctx).Create(&BuildingHiring{
		BuildingId:    b.ID,
		Payment:       payment,
		PaymentAmount: amount,
		TalentLimit:   talentLimit,
		WorkerLimit:   workerLimit,
	})
	if tx.Error == nil {
		b.setHiring(ctx)
	}
	return tx.Error
}

func (b *Building) Hiring(ctx context.Context) *BuildingHiring {
	var hiring BuildingHiring
	query := util.WithContextDb(ctx).First(&hiring, b.ID)
	if query.RecordNotFound() {
		return nil
	}
	return &hiring
}

func (h *BuildingHiring) Join(ctx context.Context, account string, apostleTokenIds []string) error {
	apostles := getApostleReversalKey(ctx, apostleTokenIds)
	for _, apostle := range apostles {
		if !strings.EqualFold(apostle.Owner, account) || apostle.Status != apostleFresh {
			return errors.New("error apostle")
		}
	}
	txn := util.DbBegin(ctx)
	defer txn.Rollback()

	var insertRecords []interface{}
	for _, tokenId := range apostleTokenIds {
		insertRecords = append(insertRecords, BuildingWorker{BuildingId: h.BuildingId, ApostleTokenId: tokenId})
	}
	if err := gormbulk.BulkInsert(txn.DB, insertRecords, 3000); err != nil {
		return err
	}

	txn.Model(Building{}).UpdateColumn(map[string]interface{}{
		"construction_duration": gorm.Expr("construction_sub + ?", len(apostleTokenIds)*perWorkerSubTime),
	})
	query := txn.Model(Apostle{}).Where("token_id in (?)", apostleTokenIds).Where("status =?", apostleFresh).Update(Apostle{Status: apostleBuild})
	if int(query.RowsAffected) != len(apostleTokenIds) {
		return errors.New("join hiring fail")
	}

	txn.DbCommit()
	return nil
}

func (b *Building) ToggleHiring(ctx context.Context) {
	util.WithContextDb(ctx).Model(b).UpdateColumn(map[string]bool{"hiring_status": !b.HiringStatus})
}

func (h *BuildingHiring) Update(ctx context.Context, payment string, amount decimal.Decimal, talentLimit, workerLimit uint) error {
	workers := h.Workers(ctx)
	if int(workerLimit) < len(workers) {
		return errors.New("error worker limit")
	}
	query := util.WithContextDb(ctx).Model(h).UpdateColumn(map[string]interface{}{
		"payment": payment, "payment_amount": amount, "talent_limit": talentLimit, "worker_limit": workerLimit,
	})
	return query.Error
}

func (h *BuildingHiring) Workers(ctx context.Context) (list []string) {
	util.WithContextDb(ctx).Model(BuildingWorker{}).Where("building_id = ?", h.BuildingId).Pluck("apostle_token_id", &list)
	return list
}

func (b *Building) AdminHiring(ctx context.Context) *BuildingAdminHiring {
	var hiring BuildingAdminHiring
	query := util.WithContextDb(ctx).First(&hiring, b.ID)
	if query.RecordNotFound() {
		return nil
	}
	return &hiring
}

func (b *BuildingAdminHiring) Update(ctx context.Context, payment string, amount decimal.Decimal, talentLimit, adminLimit uint, duration uint) error {
	admins := b.Admins(ctx)
	if int(adminLimit) < len(admins) {
		return errors.New("error worker limit")
	}
	query := util.WithContextDb(ctx).Model(b).UpdateColumn(map[string]interface{}{
		"payment": payment, "payment_amount": amount, "talent_limit": talentLimit, "worker_limit": adminLimit, "duration": duration,
	})
	return query.Error
}

func (b *BuildingAdminHiring) Admins(ctx context.Context) (list []string) {
	util.WithContextDb(ctx).Model(BuildingAdmin{}).Where("building_id = ?", b.BuildingId).Pluck("apostle_token_id", &list)
	return list
}

func (b *Building) CreateAdminHiring(ctx context.Context, payment string, amount decimal.Decimal, talentLimit, adminLimit uint, duration uint) error {
	tx := util.WithContextDb(ctx).Create(&BuildingAdminHiring{
		Payment:       payment,
		PaymentAmount: amount,
		TalentLimit:   talentLimit,
		AdminLimit:    adminLimit,
		Duration:      duration,
	})
	if tx.Error == nil {
		b.setAdminHiring(ctx)
	}
	return tx.Error
}

func (b *Building) ToggleAdmin(ctx context.Context) {
	util.WithContextDb(ctx).Model(b).UpdateColumn(map[string]bool{"admin_hiring_status": !b.AdminHiringStatus})
}

func (b *BuildingAdminHiring) Join(ctx context.Context, owner, account string, apostleTokenIds []string) error {
	apostles := getApostleReversalKey(ctx, apostleTokenIds)
	for _, apostle := range apostles {
		if !strings.EqualFold(apostle.Owner, account) || apostle.Status != apostleFresh {
			return errors.New("error apostle")
		}
	}

	txn := util.DbBegin(ctx)
	defer txn.Rollback()

	apostle := Apostle{Status: apostleBuildAdmin}
	isEmployment := owner != account
	if isEmployment {
		apostle.WorkerEnd = b.Duration*86400 + uint(time.Now().Unix())
	}
	txn.Model(Apostle{}).Where("token_id in (?)", apostleTokenIds).Update()

	var insertRecords []interface{}
	for _, tokenId := range apostleTokenIds {
		insertRecords = append(insertRecords, BuildingAdmin{BuildingId: b.BuildingId, ApostleTokenId: tokenId, IsEmployment: isEmployment})
	}
	if err := gormbulk.BulkInsert(txn.DB, insertRecords, 3000); err != nil {
		return err
	}
	txn.DbCommit()
	return nil
}

func (b *BuildingAdminHiring) WorkEnds(ctx context.Context, account string, tokenId string) error {
	apostle := GetApostleByTokenId(ctx, tokenId)
	if apostle == nil || !strings.EqualFold(apostle.Owner, account) || apostle.Status != apostleBuildAdmin {
		return errors.New("error apostle")
	}
	txn := util.DbBegin(ctx)
	defer txn.Rollback()
	txn.Model(Apostle{}).Where("token_id  = ?", tokenId).Update(Apostle{Status: apostleFresh})
	if query := txn.Where("apostle_token_id = ? ", tokenId).Delete(&BuildingAdmin{}); query.Error != nil {
		return query.Error
	}
	txn.DbCommit()
	return nil
}

func (b *Building) freeWorker(ctx context.Context) error {
	var list []string
	util.WithContextDb(ctx).Model(BuildingWorker{}).Where("building_id = ?", b.ID).Pluck("apostle_token_id", &list)
	txn := util.DbBegin(ctx)
	defer txn.Rollback()
	txn.Model(Apostle{}).Where("token_id in (?)", list).Update(Apostle{Status: apostleFresh})
	util.WithContextDb(ctx).Where("building_id = ?", b.ID).Delete(BuildingWorker{})
	txn.DbCommit()
	return nil
}

func (b *Building) renderBuildingApostle(ctx context.Context, apostleTokenIds []string) (list []BuildingApostle) {
	apostles := getApostleReversalKey(ctx, apostleTokenIds)
	for _, apostle := range apostles {
		list = append(list, BuildingApostle{
			TokenId:        apostle.TokenId,
			ApostlePicture: apostle.ApostlePicture,
			IsEmployment:   apostle.Owner != b.Address,
			TalentPoint:    88, // TODO
			TokenIndex:     getTokenIndexFromTokenId(apostle.TokenId),
		})
	}
	return list
}

func (d *Drawing) BuildingNearLand(ctx context.Context, tokenId string) ([]*Land, error) {
	land := getLand(ctx, tokenId)
	if land == nil || land.BuildingId > 0 {
		return nil, errors.New("error land tokenId or this land already building")
	}
	coords := checkLandNear(land.Lon, land.Lat, d.Index)

	xRange := []int{-112, 68}
	yRange := []int{-22, 22}
	if land.District == 2 {
		xRange = []int{68, 112}
		yRange = []int{-22, 22}
	}
	var lands []*Land
	for _, coord := range coords {
		coordSlice := strings.Split(coord, ",")
		x := util.StringToInt(coordSlice[0])
		y := util.StringToInt(coordSlice[1])
		if x < xRange[0] || x > xRange[1] || y < yRange[0] || y > yRange[1] {
			log.Debug("CreateBuilding info. x=%d. y=%d. xRange=%v; yRange=%v", x, y, xRange, yRange)
			return nil, errors.New("land choose out of range")
		}
		land := GetLandByCoord(ctx, x, y)
		if land.Status != landFresh || land.BuildingId > 0 {
			return nil, errors.New("error land tokenId or this land already building")
		}
		lands = append(lands, land)
	}
	return lands, nil
}
