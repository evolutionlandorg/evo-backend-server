package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"

	randomdata "github.com/Pallinder/go-randomdata"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type Apostle struct {
	ID             uint `gorm:"primary_key"`
	CreatedAt      time.Time
	Owner          string      `json:"owner"`
	Status         string      `json:"status"` // fresh空闲的,onsell出售中,fertility配种中,rent求职中,已雇佣hiring，working工作中,birth 诞生中
	MemberId       uint        `json:"member_id"`
	Introduction   string      `json:"introduction" sql:"type:text;"`
	District       int         `json:"district"`
	TokenId        string      `sql:"default: null" json:"token_id"`
	Chain          string      `json:"chain"`
	TokenIndex     int         `json:"token_index"`
	ApostlePicture string      `json:"apostle_picture"`
	Genes          string      `json:"genes"`
	Gen            int         `json:"gen"`
	Name           string      `json:"name"`
	Attributes     []Attribute `gorm:"many2many:apostle_attributes;"`
	Gender         string      `json:"gender"`
	Father         string      `json:"father"`
	Mother         string      `json:"mother"`
	ColdDown       int         `json:"cold_down"`
	ColdDownEnd    int         `json:"cold_down_end"`
	Duration       int         `json:"birth_duration"`
	BirthTime      int         `json:"birth_time"`
	HabergerMode   int         `json:"haberger_mode"`
	OriginId       uint        `json:"origin_id"` // 原主人，因为onsell,fertility,rent,working,hiring后owner就发生变化
	OriginAddress  string      `json:"origin_address"`
	TransTime      int         `json:"trans_time"`
	Talent         string      `json:"talent"`
	BindPetCount   uint        `json:"bind_pet_count" sql:"default: 0"`
	WorkerEnd      uint        `json:"worker_end"`
	Genesis        bool        `json:"genesis"`
	Occupational   string      `json:"occupational"`
}

type ApostleJson struct {
	Id              uint               `json:"id"`
	CreatedAt       time.Time          `json:"created_at"`
	Owner           string             `json:"owner"`
	OwnerName       string             `json:"owner_name"`
	OriginOwner     string             `json:"origin_owner"`
	OriginOwnerName string             `json:"origin_owner_name"`
	Status          string             `json:"apostle_status"`
	Introduction    string             `json:"introduction" sql:"type:text;"`
	TokenId         string             `json:"token_id"`
	TokenIndex      int                `json:"token_index"`
	ApostlePicture  string             `json:"apostle_picture"`
	Gen             int                `json:"gen"`
	Genes           string             `json:"genes"`
	Name            string             `json:"name"`
	Gender          string             `json:"gender"`
	ColdDown        int                `json:"cold_down"`
	Duration        int                `json:"duration"`
	ApostleTalent   *ApostleTalentJson `json:"apostle_talent"`
	CurrentPrice    decimal.Decimal    `json:"current_price"`
	ColdDownEnd     int                `json:"cold_down_end"`
	OriginId        uint               `json:"origin_id"`
	MemberId        uint               `json:"member_id"`
	HabergerMode    int                `json:"haberger_mode"`
	SystemSell      bool               `json:"system_sell"`
	IsAlien         bool               `json:"is_alien"`
	BirthTime       int                `json:"birth_time"`
	MineLastBid     bool               `json:"mine_last_bid"`
	HasBid          bool               `json:"has_bid"`
	BirthFee        decimal.Decimal    `json:"birth_fee"`
	Mother          string             `json:"mother"`
	WorkingStatus   string             `json:"working_status"`
	Pets            *ApostlePetJson    `json:"pets"`
	BindPetCount    uint               `json:"bind_pet_count"`
	WorkerEnd       uint               `json:"worker_end"`
	InAdventure     bool               `json:"in_adventure"`
	Occupational    string             `json:"occupational"`
	Equipments      []Equipment        `json:"equipments"`
	Token           *util.Token        `json:"token"`
}

type ApostleDetail struct {
	ApostleJson
	ApostleParents
	ApostleChildren []ApostleDisplaySample     `json:"apostle_children"`
	Auction         *ApostleAuctionSample      `json:"auction"`
	Working         *ApostleWorkInfo           `json:"working"`
	Attributes      map[string]AttributeJson   `json:"attributes"`
	DigStrength     map[string]decimal.Decimal `json:"dig_strength"`
}

type ApostleParents struct {
	ApostleFather ApostleDisplaySample `json:"apostle_father"`
	ApostleMother ApostleDisplaySample `json:"apostle_mother"`
}

type ApostleDisplaySample struct {
	TokenId        string `json:"token_id"`
	TokenIndex     int    `json:"token_index"`
	ApostlePicture string `json:"apostle_picture"`
	Mother         string `json:"apostle_mother"`
}

type ApostleSample struct {
	ApostlePicture string `json:"apostle_picture"`
	Name           string `json:"name"`
	Slot           int    `json:"slot"`
	ApostleTokenId string `json:"apostle_token_id"`
}

type ApostleQuery struct {
	WhereQuery struct {
		Owner         string `table_name:"apostles" json:"owner"`
		District      int    `table_name:"apostles" json:"district"`
		Status        string `table_name:"apostles" json:"status"`
		Gen           int    `table_name:"apostles" json:"gen"`
		OriginId      uint   `table_name:"apostles" json:"origin_id"`
		TokenIndex    int    `table_name:"apostles" json:"token_index"`
		Gender        string `table_name:"apostles" json:"gender"`
		OriginAddress string `table_name:"apostles" json:"origin_address"`
	}
	MultiFilter struct {
		Element []string
		Price   map[string]string
		Gen     map[string]string
		Talent  map[string]string
	}
	TokenId    []string
	MyLastBid  []string
	HasBid     []string
	Tokens     map[string]*util.Token
	IncludeId  []int
	ExcludeId  []int
	Row        int
	Page       int
	Filter     string
	OrderField string
	Order      string
	Display    string
	Chain      string
}

var (
	ApostlesOrder = []string{"price", "gen", "id", "token_index", "atk", "crit", "def", "hp_limit", "occupational", "mining_power"}
)

const (
	ApostleGenderMale   = "male"
	ApostleGenderFemale = "female"
)

func (ap *Apostle) New(db *util.GormDB) error {
	ap.OriginId = 0
	ap.ColdDownEnd = 0
	ap.HabergerMode = 0
	ap.TransTime = int(time.Now().Unix())
	result := db.Create(&ap)
	return result.Error
}

func GetApostlePicture(genes, chain string, tokenIndex int) string {
	if tokenIndex <= 1 {
		return util.Evo.GenesisApostlePicture[chain]
	}
	u, err := url.Parse(util.ApiServerHost)
	if err != nil {
		return ""
	}
	u.Path = fmt.Sprintf("/apostle/%s.png", util.U256(genes).String())
	return u.String()
}

func (ap *Apostle) ApostleTalent(ctx context.Context) *ApostleTalentJson {
	db := util.WithContextDb(ctx)
	var at ApostleTalentJson
	if q := db.Table("apostle_talents").Where("token_id = ?", ap.TokenId).Scan(&at); q.RecordNotFound() {
		return nil
	}
	return &at
}

func (ap *Apostle) TransferOwner(db *util.GormDB, owner, status, originAddress string, originId uint, chain string) error {
	updateField := map[string]interface{}{"owner": owner, "status": status, "origin_id": originId, "origin_address": originAddress, "genesis": false}
	if strings.EqualFold(originAddress, util.GetContractAddress("Gen0", chain)) {
		updateField["genesis"] = true
	}
	updateField["trans_time"] = int(time.Now().Unix())
	query := db.Model(&ap).UpdateColumn(updateField)
	if query.RowsAffected > 0 {
		// 转移装备的 origin_address
		_ = SetEquipmentOriginOwnerByApostleId(db, ap.TokenId, util.TrueOrElse(originAddress != "", originAddress, owner))
	}
	return query.Error
}

func (ap *Apostle) AsJson(ctx context.Context) *ApostleDetail {
	if ap == nil {
		return nil
	}
	auction := ap.getApostleAuctionJson(ctx)
	chain := GetChainByTokenId(ap.TokenId)
	member := GetMemberByAddress(ctx, ap.Owner, chain)
	var ownerName string
	if member != nil {
		ownerName = member.Name
	}
	apostleJson := ApostleJson{
		Id:             ap.ID,
		MemberId:       ap.MemberId,
		CreatedAt:      ap.CreatedAt,
		Owner:          ap.Owner,
		OwnerName:      ownerName,
		Name:           ap.Name,
		Status:         ap.Status,
		Introduction:   ap.Introduction,
		TokenId:        ap.TokenId,
		TokenIndex:     ap.TokenIndex,
		ApostlePicture: ap.ApostlePicture,
		Gen:            ap.Gen,
		Genes:          ap.Genes,
		Gender:         ap.Gender,
		ColdDown:       ap.ColdDown,
		ColdDownEnd:    ap.ColdDownEnd,
		OriginId:       ap.OriginId,
		BirthTime:      ap.BirthTime,
		CurrentPrice:   auction.CurrentPrice,
		Duration:       ap.Duration,
		HabergerMode:   ap.HabergerMode,
		BirthFee:       apostleBirthFee,
		Mother:         ap.Mother,
		BindPetCount:   ap.BindPetCount,
		Occupational:   ap.Occupational,
	}
	var apostles []ApostleJson
	apostles = append(apostles, apostleJson)
	apostles = renderingApostle(ctx, chain, apostles)
	priceMap, tokenMap := ApostlePriceCache(ctx, chain)
	if price, ok := priceMap[apostles[0].TokenId]; ok {
		apostles[0].CurrentPrice = price
	}
	apostles[0].Token = tokenMap[apostles[0].TokenId]
	auction.Token = apostles[0].Token
	apostles = renderApostleOwner(ctx, apostles)
	working := ap.ApostleRentInfo(ctx)
	ad := ApostleDetail{
		apostles[0],
		ap.Parent(ctx),
		ap.Children(ctx),
		auction,
		working,
		ap.getAttributes(ctx),
		apostles[0].digStrengthBase(),
	}
	return &ad
}

func (ap *Apostle) Parent(ctx context.Context) ApostleParents {
	var parents ApostleParents
	if ap.Gen == 0 {
		return parents
	}
	var parentsIds []string
	parentsIds = append(parentsIds, ap.Mother)
	parentsIds = append(parentsIds, ap.Father)
	parentsApostle := getApostleReversalKey(ctx, parentsIds)
	parents.ApostleFather = ApostleDisplaySample{parentsApostle[ap.Father].TokenId,
		parentsApostle[ap.Father].TokenIndex, parentsApostle[ap.Father].ApostlePicture, ""}
	parents.ApostleMother = ApostleDisplaySample{parentsApostle[ap.Mother].TokenId,
		parentsApostle[ap.Mother].TokenIndex, parentsApostle[ap.Mother].ApostlePicture, ""}

	return parents
}

func (ap *Apostle) Children(ctx context.Context) []ApostleDisplaySample {
	db := util.WithContextDb(ctx)
	var children []ApostleDisplaySample
	db.Model(&Apostle{}).Where("father = ?", ap.TokenId).Or("mother = ?", ap.TokenId).Order("id asc").Scan(&children)
	return children
}

func (ap *Apostle) getApostleAuctionJson(ctx context.Context) *ApostleAuctionSample {
	var as ApostleAuctionSample
	chain := GetChainByTokenId(ap.TokenId)
	as.CurrentTime = time.Now().Unix()
	if ap.Status == apostleOnsell {
		auc := GetApostleAuction(ctx, ap.TokenId, AuctionGoing)
		if auc != nil {
			as.CurrentPrice = auc.CurrentPriceFromChain()
			as.StartAt = auc.StartAt
			as.Duration = auc.Duration
			as.Seller = auc.Seller
			as.StartPrice = auc.StartPrice
			as.EndPrice = auc.EndPrice
			as.ClaimWaiting = int(apostleClaimTime(ap.District))
			as.LastPrice = auc.LastPrice
			as.LastBidStart = auc.LastBidStart
			as.History = auctionHistoryList(ctx, auc.ID, AssetApostle, ap.TokenId)
			member := GetMemberByAddress(ctx, auc.Seller, chain)
			if member != nil {
				as.SellerName = member.Name
			}
			if as.LastBidStart != 0 && time.Now().Unix()-int64(as.LastBidStart) > int64(as.ClaimWaiting) {
				as.Status = AuctionClaimed
				as.Winner = auc.LastBidder
				winner := GetMemberByAddress(ctx, auc.LastBidder, chain)
				if winner != nil {
					as.WinnerName = winner.Name
				}
			}
		}
	} else if ap.Status == apostleFertility {
		af := GetApostleFertility(ctx, ap.TokenId, AuctionGoing)
		if af != nil {
			as.Duration = af.Duration
			as.StartAt = af.StartAt
			as.CurrentPrice = af.CurrentPrice()
			as.StartPrice = af.StartPrice
			as.EndPrice = af.EndPrice
			as.Seller = af.Seller
			as.SellerName = af.Seller
			member := GetMemberByAddress(ctx, af.Seller, chain)
			if member != nil {
				as.SellerName = member.Name
			}
		}
	} else if ap.Status == apostleRent {
		ar := GetApostleWork(ctx, ap.TokenId, AuctionGoing)
		if ar != nil {
			as.Duration = ar.Duration
			as.CurrentPrice = ar.Price
			as.Seller = ar.Seller
			member := GetMemberByAddress(ctx, ar.Seller, chain)
			if member != nil {
				as.SellerName = member.Name
			}
		}
	}
	return &as
}

func (ap *Apostle) startHabergerMode(db *util.GormDB) {
	if ap.HabergerMode == 0 {
		db.Model(&Apostle{}).Where("id = ?", ap.ID).UpdateColumn(Apostle{HabergerMode: 1})
	}
}

func (ap *Apostle) initApostleSvg() {
	if ap.Genes != "" {
		u, err := url.Parse(util.ApostlePictureServer)
		if err != nil {
			log.Fatalf("parse url error: %s", err)
		}
		u.Path = fmt.Sprintf("/%s.svg", util.U256(ap.Genes).String())
		go util.HttpGet(u.String())
	}
}

func (apq *ApostleQuery) Apostles(ctx context.Context, occupational []string) (*[]ApostleJson, int) {
	db := util.WithContextDb(ctx)
	var (
		apostles []ApostleJson
		count    int
		query    *gorm.DB
	)
	ignoreTokenID := []string{"fertility", "rent", "my", "fresh", "canWorking", "unbind", "mine"}
	if apq.Filter != "" && !util.StringInSlice(apq.Filter, ignoreTokenID) && apq.Display != "all" {
		query = db.Table("apostles").Where(apq.WhereQuery).Where("apostles.token_id in (?)", apq.TokenId).
			Order("apostles.id desc")
	} else {
		switch apq.Filter {
		case "my":
			owner := apq.WhereQuery.Owner
			apq.WhereQuery.Owner = ""
			query = db.Table("apostles").Where(apq.WhereQuery).
				Where("owner=? or (status=? and origin_address =?) or (status=? and origin_address =?) or (status=? and origin_address =?)", owner, apostleWorking, owner, apostleHiring, owner, apostlePve, owner).
				Order("apostles.trans_time desc")
		case "unbind":
			query = db.Table("apostles").Where(apq.WhereQuery).Where("origin_address =? and bind_pet_count< 1 and status!=?", "", apostleBirth).Order("apostles.trans_time desc")
		case "mine":
			owner := apq.WhereQuery.Owner
			apq.WhereQuery.Owner = ""
			query = db.Table("apostles").Where(apq.WhereQuery).Where("owner = ? or origin_address =?", owner, owner).
				Order("apostles.trans_time desc")
		case "listing":
			query = db.Table("apostles").Where(apq.WhereQuery).Where("apostles.status in (?)", []string{"onsell", "fertility", "rent"})
		case "employment":
			query = db.Table("apostles").Where(apq.WhereQuery).Where("apostles.status in (?)", []string{apostleBuildAdmin, apostleWorking})
		case "onsell":
			query = db.Table("apostles").Where(apq.WhereQuery).Order("genesis desc")
		default:
			query = db.Table("apostles").Where(apq.WhereQuery).Where("status !=?", apostleBirth).Order("apostles.id desc")
		}
	}
	if len(occupational) != 0 {
		query = query.Where("occupational IN (?)", occupational)
	}
	// first need filter
	priceMap, tokenMap := ApostlePriceCache(ctx, apq.Chain)
	tokenQuery, tokenIndex, needFilter := apq.multiFilterApostle(ctx, priceMap)
	for _, w := range tokenQuery {
		query = query.Where(w)
	}
	if needFilter {
		if len(tokenIndex) == 0 {
			return nil, 0
		}
		query = query.Where("token_index in (?)", tokenIndex)
	}

	query = query.Scan(&apostles)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil, 0
	}

	for index, apostle := range apostles {
		apostles[index].CurrentPrice = priceMap[apostle.TokenId]
		apostles[index].Token = tokenMap[apostle.TokenId]
	}

	apostles = apq.filterApostle(ctx, apostles)
	count = len(apostles)
	if apq.Page*apq.Row > count || count == 0 {
		return nil, 0
	}
	apostles = apq.sortApostles(ctx, apostles)
	maxLength := (apq.Page + 1) * apq.Row
	if maxLength > count {
		maxLength = count
	}
	apostles = apostles[apq.Page*apq.Row : maxLength]
	apostles = renderingApostle(ctx, apq.Chain, apostles)
	apostles = renderApostleOwner(ctx, apostles)
	if apq.Tokens == nil {
		apq.Tokens = make(map[string]*util.Token)
	}
	for i, v := range apostles {
		if util.StringInSlice(v.TokenId, apq.MyLastBid) {
			apostles[i].MineLastBid = true
		}
		if util.StringInSlice(v.TokenId, apq.HasBid) {
			apostles[i].HasBid = true
		}
		if apostles[i].Token == nil {
			apostles[i].Token = apq.Tokens[v.TokenId]
		}
		apostles[i].ApostlePicture = GetApostlePicture(v.Genes, apq.Chain, v.TokenIndex)
	}

	return &apostles, count
}

func (apq *ApostleQuery) AllApostles(ctx context.Context, occupational []string) (*[]ApostleJson, int) {
	db := util.WithContextDb(ctx)
	var (
		apostles []ApostleJson
		count    int
	)

	wheres, value := util.StructToSql(apq.WhereQuery)
	query := db.Table("apostles").Where(strings.Join(wheres, " AND "), value...)
	db = db.Table("apostles").Where(strings.Join(wheres, " AND "), value...)
	if len(occupational) != 0 {
		query = query.Where("occupational IN (?)", occupational)
		db = db.Where("occupational IN (?)", occupational)
	}

	switch strings.ToLower(apq.OrderField) {
	case "mining_power", "atk", "hp_limit":
		query = query.Joins("INNER JOIN apostle_talents as a ON a.token_id COLLATE utf8mb4_general_ci = apostles.token_id").Select("apostles.*")
		query = query.Order(fmt.Sprintf("a.%s %s", apq.OrderField, apq.Order))
	case "gen", "token_index":
		query = query.Order(fmt.Sprintf("apostles.%s %s", apq.OrderField, apq.Order))
	case "price":
		query = query.Joins("LEFT JOIN auction_apostles as a ON a.token_id = apostles.token_id")
		query = query.Order(fmt.Sprintf("a.start_%s %s", apq.OrderField, apq.Order))
	}
	query.Where("apostles.status !=?", apostleBirth).Offset(apq.Page * apq.Row).Limit(apq.Row).Scan(&apostles)
	db.Where("apostles.status !=?", apostleBirth).Count(&count)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil, 0
	}
	apostles = renderingApostle(ctx, apq.Chain, apostles)
	apostles = renderApostleOwner(ctx, apostles)
	if apq.Tokens == nil {
		apq.Tokens = make(map[string]*util.Token)
	}
	for index, apostle := range apostles {
		if apostle.Token == nil {
			apostles[index].Token = apq.Tokens[apostle.TokenId]
		}
		apostles[index].ApostlePicture = GetApostlePicture(apostle.Genes, apq.Chain, apostle.TokenIndex)
	}
	return &apostles, count
}

func (apq *ApostleQuery) sortApostles(ctx context.Context, apostles []ApostleJson) []ApostleJson {
	if util.StringInSlice(apq.OrderField, ApostlesOrder) {
		if util.StringInSlice(apq.OrderField, []string{"atk", "def", "hp_limit", "crit", "mining_power"}) {
			var ids []uint
			for _, v := range apostles {
				ids = append(ids, v.Id)
			}
			talents := getApostleTalentReversalKey(ctx, ids)
			for i, v := range apostles {
				talent := talents[v.Id]
				apostles[i].ApostleTalent = &talent.ApostleTalentJson
			}
		}
		sort.Slice(apostles[:], func(i, j int) bool {
			if apq.Order == "desc" {
				i, j = j, i
			}
			switch apq.OrderField {
			case "price":
				return apostles[i].CurrentPrice.LessThan(apostles[j].CurrentPrice)
			case "atk", "def", "hp_limit", "crit", "mining_power":

				if apostles[i].ApostleTalent == nil {
					return true
				}
				orderFiled := apq.OrderField
				if orderFiled == "hp_limit" {
					orderFiled = "HPLimit"
				} else if orderFiled == "mining_power" {
					orderFiled = "MiningPower"
				} else {
					orderFiled = strings.ToUpper(orderFiled)
				}
				valuePrev, _ := getStringValueByFieldName(apostles[i].ApostleTalent, orderFiled)
				valueNext, _ := getStringValueByFieldName(apostles[j].ApostleTalent, orderFiled)
				return decimal.RequireFromString(valuePrev).LessThan(decimal.RequireFromString(valueNext))
			case "occupational":
				return apostles[i].Occupational == ""
			default:
				valuePrev, _ := getStringValueByFieldName(apostles[i], util.CamelString(apq.OrderField))
				valueNext, _ := getStringValueByFieldName(apostles[j], util.CamelString(apq.OrderField))
				return util.StringToInt(valuePrev) < util.StringToInt(valueNext)
			}
		})
	}
	return apostles
}

func (apq *ApostleQuery) filterApostle(ctx context.Context, apostles []ApostleJson) []ApostleJson {
	if len(apq.IncludeId) > 0 {
		var newApostles []ApostleJson
		for _, v := range apostles {
			if util.IntInSlice(int(v.Id), apq.IncludeId) {
				newApostles = append(newApostles, v)
			}
		}
		apostles = newApostles
	}
	if len(apq.ExcludeId) > 0 {
		var newApostles []ApostleJson
		for _, v := range apostles {
			if !util.IntInSlice(int(v.Id), apq.ExcludeId) {
				newApostles = append(newApostles, v)
			}
		}
		apostles = newApostles
	}
	switch apq.Filter {
	case "canWorking":
		apostles = filterCanWorking(ctx, apq.WhereQuery.Owner, apostles)
	case "onsell":
		if apq.Display == "all" {
			apostles = filterCanSell(ctx, apostles, apq.Chain)
		}
	}
	return apostles
}

func (apq *ApostleQuery) multiFilterApostle(ctx context.Context, priceMap map[string]decimal.Decimal) (tokenQuery []string, apostleTokenIndex []int64, needFilter bool) {
	var compare PriceCompare
	multiFilter := apq.MultiFilter

	var apostleElementId []string
	tokenPrefix := util.Evo.ApostlePrefix[apq.Chain] + "%"
	if (len(multiFilter.Element) > 0 && len(multiFilter.Element) < 5) || len(multiFilter.Talent) > 0 {
		query := util.WithContextDb(ctx).Model(ApostleTalent{}).Where("token_id LIKE ?", tokenPrefix)
		for _, element := range multiFilter.Element {
			switch element {
			case currencyGold:
				query = query.Where("element_gold > 0")
			case currencyWood:
				query = query.Where("element_wood > 0")
			case currencyWater:
				query = query.Where("element_water > 0")
			case currencyFire:
				query = query.Where("element_fire > 0")
			case currencySoil:
				query = query.Where("element_soil > 0")
			}
		}
		for talent, value := range multiFilter.Talent {
			query = query.Where(fmt.Sprintf("%s >= ?", talent), value)
		}
		query.Pluck("token_id", &apostleElementId)
		needFilter = true
	}

	if len(multiFilter.Price) > 0 {
		util.UnmarshalAny(&compare, multiFilter.Price)
		checkPrice := func(priceMap map[string]decimal.Decimal) {
			for tokenId, price := range priceMap {
				if compare.Gte != nil && price.LessThan(*compare.Gte) {
					continue
				}
				if compare.Lte != nil && price.GreaterThan(*compare.Lte) {
					continue
				}
				apostleElementId = append(apostleElementId, tokenId)
			}
		}
		checkPrice(priceMap)
		needFilter = true
	}

	if len(multiFilter.Gen) > 0 {
		if get, ok := multiFilter.Gen["gte"]; ok {
			tokenQuery = append(tokenQuery, fmt.Sprintf("gen >= %d ", util.StringToInt(get)))
		}
		if lte, ok := multiFilter.Gen["lte"]; ok {
			tokenQuery = append(tokenQuery, fmt.Sprintf("gen <= %d ", util.StringToInt(lte)))
		}
	}
	for _, tokenId := range apostleElementId {
		apostleTokenIndex = append(apostleTokenIndex, new(big.Int).Mod(util.U256(tokenId), big.NewInt(65536)).Int64())
	}
	return
}

func (ap *ApostleJson) digStrengthBase() map[string]decimal.Decimal {
	digStrengthBase := decimal.Zero
	talent := ap.ApostleTalent
	if *talent.Potential < 100 {
		digStrengthBase = decimal.New(int64(*talent.Strength), 0).Mul(decimal.New(int64(*talent.Agile), 0)).Div(decimal.New(7, 0)).Div(decimal.New(int64(*talent.Potential), 0))
	}
	if *talent.Potential >= 100 {
		digStrengthBase = decimal.New(int64(*talent.Strength), 0).Mul(decimal.New(int64(*talent.Agile), 0)).Div(decimal.New(8, 0)).Div(decimal.New(int64(*talent.Potential), 0))
	}
	return map[string]decimal.Decimal{"min": digStrengthBase.Mul(decimal.NewFromFloat(1.03)).Round(3), "max": digStrengthBase.Mul(decimal.NewFromFloat(1.766)).Round(3)}
}

// ApostleCallback 使徒出生/怀孕回调函数
func (ec *EthTransactionCallback) ApostleCallback(ctx context.Context) (err error) {
	if getTransactionDeal(ctx, ec.Tx, "apostle") != nil {
		return errors.New("tx exist")
	}
	db := util.DbBegin(ctx)
	defer db.DbRollback()

	chain := ec.Receipt.ChainSource

	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("apostle", chain)) {
			eventName := util.AddHex(log.Topics[0])
			logSlice := util.LogAnalysis(log.Data)
			switch eventName {
			case services.AbiEncodingMethod("Birth(address,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256)"):
				owner := util.AddHex(util.TrimHex(log.Topics[1])[24:64], chain)
				coolDown := util.StringToInt(util.U256(logSlice[5]).String())
				Gen := util.StringToInt(util.U256(logSlice[6]).String())
				birthTime := util.StringToInt(util.U256(logSlice[7]).String())
				err = createApostles(ctx, db, owner, logSlice[0], logSlice[1], logSlice[2], logSlice[3], logSlice[4], Gen, coolDown, birthTime)

			case services.AbiEncodingMethod("Pregnant(uint256,uint256,uint256,uint256,uint256,uint256)"):
				MatronCoolDownEnd := util.StringToInt(util.U256(logSlice[1]).String())
				SireCoolDownEnd := util.StringToInt(util.U256(logSlice[4]).String())
				MatronCoolDown := util.StringToInt(util.U256(logSlice[2]).String())
				SireCoolDown := util.StringToInt(util.U256(logSlice[5]).String())
				err = dealApostlePregnant(ctx, db, ec.Tx, logSlice[0], logSlice[3], MatronCoolDown, SireCoolDown, MatronCoolDownEnd, SireCoolDownEnd)
			}
		}
	}

	if err != nil {
		return err
	}

	if err := ec.NewUniqueTransaction(db, "apostle"); err != nil {
		return err
	}

	db.DbCommit()
	return nil
}

func GetApostleByTokenId(ctx context.Context, tokenId string) *Apostle {
	db := util.WithContextDb(ctx)
	var apostle Apostle
	query := db.Model(&apostle).Where("token_id = ?", tokenId).Scan(&apostle)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	apostle.ApostlePicture = GetApostlePicture(apostle.Genes, apostle.Chain, apostle.TokenIndex)
	return &apostle
}

// GetApostleMiningPower 计算使徒的挖矿力
// strength * agility / ((7 + potential / 100) * potential)
func GetApostleMiningPower(strength, agile, potential decimal.Decimal) decimal.Decimal {
	// (7 + potential / 100) * potential)
	b := potential.Div(decimal.NewFromInt(100)).Add(decimal.NewFromInt(7)).Mul(potential)
	// strength * agility / b
	if b.IsZero() {
		return decimal.NewFromInt(0)
	}
	return strength.Mul(agile).Div(b)
}

func changeApostleOwner(db *util.GormDB, tokenId, owner, chain string) error {
	if util.SliceIndex(owner, []string{util.GetContractAddress("clockAuctionApostle", chain),
		util.GetContractAddress("apostleFertility", chain),
		util.GetContractAddress("tokenUse", chain)}, true) == -1 {
		_ = SetEquipmentOriginOwnerByApostleId(db, tokenId, owner)
	}
	query := db.Table("apostles").Where("token_id = ?", tokenId).
		UpdateColumn(map[string]interface{}{"owner": owner, "trans_time": int(time.Now().Unix())})
	return query.Error
}

// func getApostleByName(name string) *Apostle {
//	db := util.DB
//	var apostle Apostle
//	query := db.Model(&apostle).Where("name = ?", name).Scan(&apostle)
//	if query.Error != nil || query == nil || query.RecordNotFound() {
//		return nil
//	}
//	return &apostle
// }

func getApostleReversalKey(ctx context.Context, ids []string) map[string]Apostle {
	if len(ids) < 1 {
		return nil
	}
	db := util.WithContextDb(ctx)
	var apostles []Apostle
	query := db.Where("token_id in (?)", ids).Find(&apostles)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	results := make(map[string]Apostle)
	for _, v := range apostles {
		results[v.TokenId] = v
	}
	return results
}

func getApostleReversalKeyByIds(ctx context.Context, ids []uint) map[uint]Apostle {
	if len(ids) < 1 {
		return nil
	}
	db := util.WithContextDb(ctx)
	var apostles []Apostle
	query := db.Where("id in (?)", ids).Find(&apostles)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	results := make(map[uint]Apostle)
	for _, v := range apostles {
		results[v.ID] = v
	}
	return results
}

func renderingApostle(ctx context.Context, chain string, apostles []ApostleJson) []ApostleJson {
	if len(apostles) == 0 {
		return nil
	}
	// remove onFertilityIds, onRentIds
	var onsellIds, onWorking, bindPetApostleIds, apostleIds []uint
	var tokenIds []string
	var initTalents bool

	for _, v := range apostles {
		apostleIds = append(apostleIds, v.Id)
		if v.TokenId != "" {
			tokenIds = append(tokenIds, v.TokenId)
		}
		if v.BindPetCount > 0 {
			bindPetApostleIds = append(bindPetApostleIds, v.Id)
		}
		switch v.Status {
		case apostleOnsell:
			onsellIds = append(onsellIds, v.Id)
		case apostleFertility:
			// this result of append is never used, except maybe in other appends
			// onFertilityIds = append(onFertilityIds, v.Id)
		case apostleRent:
			// this result of append is never used, except maybe in other appends
			// onRentIds = append(onRentIds, v.Id)
		case apostleWorking, apostleHiring:
			onWorking = append(onWorking, v.Id)
		}
		if v.ApostleTalent != nil {
			initTalents = true
		}
	}

	onsellAuction := getAuctionApostleReversalKey(ctx, onsellIds, AuctionGoing)
	// onFertilityAuction := getApostleFertilityReversalKey(onFertilityIds, AuctionGoing)
	// onRentAuction := getApostleWorkTradeReversalKey(onRentIds, AuctionGoing)
	onWorkingAuction := getApostleWorkTradeReversalKey(ctx, onWorking, AuctionFinish)
	bindPetApostle := getApostlePetsReversalKey(ctx, bindPetApostleIds)
	var talents map[uint]ApostleTalent
	if !initTalents {
		talents = getApostleTalentReversalKey(ctx, apostleIds)
	}
	// 装备
	eqs := apostle2Equipments(ctx, tokenIds)

	for i, apostle := range apostles {
		apostles[i].IsAlien = getApostlesAlienFromGenes(apostle.Genes)
		apostles[i].BirthFee = apostleBirthFee
		apostles[i].Pets = nil
		if bindPetApostle != nil && bindPetApostle[apostle.Id].ApostleId != 0 {
			pet := bindPetApostle[apostle.Id]
			apostles[i].Pets = &pet
		}
		if apostle.Status == apostleFresh {
			apostles[i].ColdDownEnd = apostles[i].ColdDownEnd - int(time.Now().Unix())
			if apostles[i].ColdDownEnd < 0 || apostle.ColdDownEnd == 0 {
				apostles[i].ColdDownEnd = 0 // 可再次生育倒计时
			}
		}
		if apostle.Status == apostleOnsell && onsellAuction != nil {
			auc := onsellAuction[apostle.Id]
			if auc.Seller == util.GetContractAddress("Gen0", GetChainByTokenId(apostle.TokenId)) {
				apostles[i].SystemSell = true
				if auc.StartAt > int(time.Now().Unix()) {
					apostles[i].Status = apostleFresh
				}
			}
			apostles[i].CurrentPrice = auc.CurrentPrice()
			if auc.LastBidStart != 0 && time.Now().Unix()-int64(auc.LastBidStart) > apostleClaimTime(getNFTDistrict(apostle.TokenId)) {
				apostles[i].Status = AuctionClaimed
			}
			apostles[i].Token = util.Evo.GetToken(GetChainByDistrict(auc.District), auc.Currency)
		} else if apostle.Status == apostleBirth {
			// } else if apostle.Status == apostleFertility && onFertilityAuction != nil {
			// 	// af := onFertilityAuction[apostle.Id]
			// 	// apostles[i].CurrentPrice = af.CurrentPrice()
			// } else if apostle.Status == apostleRent && onRentAuction != nil {
			// 	// ar := onRentAuction[apostle.Id]
			// 	// if ar.Duration > 0 {
			// 	// 	apostles[i].CurrentPrice = ar.Price.Div(decimal.New(int64(ar.Duration/86400), 0))
			// 	}
			// } else if apostle.Status == apostleBirth {
			duration := apostle.Duration - int(time.Now().Unix())
			if duration < 0 {
				apostles[i].Status = apostleBirthUnclaimed // 出生未领取
				duration = 0
			}
			apostles[i].Duration = duration
		} else if apostle.Status == apostleHiring || apostle.Status == apostleWorking {
			aw := onWorkingAuction[apostle.Id]
			apostles[i].WorkingStatus = apostle.Status
			duration := aw.StartAt + aw.Duration - int(time.Now().Unix())
			if aw.ID != 0 && duration < 0 {
				apostles[i].Status = apostleHiringUnclaimed // 租用到期未领取
				duration = 0
			}
			apostles[i].Duration = duration
		}

		// ApostleTalent
		if j, ok := talents[apostle.Id]; ok {
			talent := j.ApostleTalentJson
			apostles[i].ApostleTalent = &talent
		}
		if eq, ok := eqs[apostle.TokenId]; ok {
			apostles[i].Equipments = eq
		}

	}
	return apostles
}

func renderApostleOwner(ctx context.Context, apostles []ApostleJson) []ApostleJson {
	for i, apostle := range apostles {
		if apostle.TokenId != "" {
			chain := GetChainByTokenId(apostle.TokenId)
			if apostle.OriginId > 0 {
				if member := GetMember(ctx, int(apostle.OriginId)); member != nil {
					apostles[i].OriginOwnerName = member.Name
					apostles[i].OriginOwner = member.GetUseAddress(chain)
				}
			} else {
				if member := GetMemberByAddress(ctx, apostle.Owner, chain); member != nil {
					apostles[i].OriginOwnerName = member.Name
					apostles[i].OriginOwner = apostle.Owner
				}
			}
		}
	}
	return apostles
}

func createApostles(ctx context.Context, db *util.GormDB, owner, tokenId, motherTokenId, fatherTokenId, genes, talents string, gen, coolDown, birthTime int) error {
	tokenIndex := getTokenIndexFromTokenId(tokenId)
	gender := getApostlesGenderFromGenes(genes)
	var father, mother string
	if util.U256(motherTokenId).String() != "0" && util.U256(fatherTokenId).String() != "0" {
		mother = motherTokenId
		father = fatherTokenId
	}
	ap := Apostle{Status: apostleFresh, Owner: owner, TokenId: tokenId, TokenIndex: tokenIndex,
		Genes: genes, Gen: gen, ColdDown: coolDown, Gender: gender, Father: father, Mother: mother, BirthTime: birthTime, Duration: 0, ColdDownEnd: 0}
	ap.Name = randomUniqueName(ap.Gender)
	ap.Introduction = randomUniqueApostleDesc(ap.Name)
	ap.ApostlePicture = GetApostlePicture(genes, ap.Chain, ap.TokenIndex)
	ap.District = getNFTDistrict(ap.TokenId)
	ap.Chain = GetChainByTokenId(ap.TokenId)
	member := GetOrCreateMemberByAddress(ctx, owner, GetChainByTokenId(tokenId))
	if member != nil {
		ap.MemberId = member.ID
	}
	var apostleBaby Apostle
	if father != "" && mother != "" {
		db.Model(&Apostle{}).Where("status = ?", apostleBirth).Where("father = ?", father).Where("mother = ?", mother).First(&apostleBaby)
	}
	if apostleBaby.ID != 0 {
		db.Model(&apostleBaby).UpdateColumn(map[string]interface{}{"status": apostleFresh, "owner": owner, "token_id": tokenId, "chain": GetChainByTokenId(tokenId), "token_index": tokenIndex,
			"genes": genes, "gen": gen, "cold_down": coolDown, "gender": gender, "birth_time": birthTime, "duration": 0, "cold_down_end": 0,
			"name": ap.Name, "introduction": ap.Introduction, "apostle_picture": ap.ApostlePicture, "district": getNFTDistrict(tokenId)})
		ap = apostleBaby
	} else {
		if err := ap.New(db); err != nil {
			return errors.New("Apostle create error : " + err.Error())
		}
	}
	if err := ap.createApostleTalentFromTalent(db, talents); err != nil {
		return errors.New("Apostle create talent error : " + err.Error())
	}
	ap.initApostleSvg()
	ap.createAttributeFromGene(ctx)
	refreshOpenSeaMetadata(tokenId)
	return nil
}

func getApostlesGenderFromGenes(genes string) string {
	b := util.EncodeU256(genes)
	ls := new(big.Int).Lsh(big.NewInt(1), 241)
	ts := new(big.Int).And(b, ls)
	if ts.Sign() == 0 {
		return ApostleGenderFemale
	}
	return ApostleGenderMale
}

func getApostlesAlienFromGenes(genes string) bool {
	if genes == "" {
		return false
	}
	b := util.EncodeU256(genes)
	ls := new(big.Int).Lsh(big.NewInt(3), 242)
	ts := new(big.Int).And(b, ls)
	return ts.Sign() != 0
}

func randomUniqueName(gender string) string {
	var name string
	if gender == ApostleGenderMale {
		name = randomdata.FullName(0)
	} else {
		name = randomdata.FullName(1)
	}
	return util.CamelString(name)

}

func randomUniqueApostleDesc(apostleName string) string {
	var apostleDescJson map[string][]string
	b, _ := os.ReadFile("config/apostle_desc.json")
	_ = json.Unmarshal(b, &apostleDescJson)
	hello := apostleDescJson["hello"]
	name := apostleDescJson["name"]
	sentence := apostleDescJson["sentence"]
	says := apostleDescJson["says"]
	bye := apostleDescJson["bye"]
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	desc := fmt.Sprintf("%s %s %s %s %s", hello[rnd.Intn(len(hello))], name[rnd.Intn(len(name))], sentence[rnd.Intn(len(sentence))], says[rnd.Intn(len(says))], bye[rnd.Intn(len(bye))])
	desc = strings.Replace(desc, "his_name", apostleName, -1)
	return desc

}

// token_id && 2 ** 128 -1
func getTokenIndexFromTokenId(tokenId string) int {
	b := util.EncodeU256(tokenId)
	ls := new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)
	ms := new(big.Int).Sub(ls, big.NewInt(1))
	ts := new(big.Int).And(b, ms)
	return int(ts.Int64())
}

func ApostlePriceCache(ctx context.Context, chain string) (map[string]decimal.Decimal, map[string]*util.Token) {
	key := fmt.Sprintf("apsotle-price:%s", chain)
	apostlePrice := make(map[string]decimal.Decimal)
	tokenMap := make(map[string]*util.Token)
	if cache := util.GetCache(ctx, key); len(cache) == 0 {
		_, _, auctionPrice, tokenMaps := OnsellApostleList(ctx, "", chain, nil)
		siringAuctionPrice, siringTokenMaps := SiringApostlePriceList(ctx, chain)
		workerAuctionPrice, workTokenMaps := WorkerPriceList(ctx, chain)

		checkPrice := func(priceMap map[string]decimal.Decimal, t map[string]*util.Token) {
			for tokenId, price := range priceMap {
				apostlePrice[tokenId] = price
			}
			for i := range t {
				tokenMap[i] = t[i]
			}
		}

		checkPrice(auctionPrice, tokenMaps)
		checkPrice(siringAuctionPrice, siringTokenMaps)
		checkPrice(workerAuctionPrice, workTokenMaps)
		bp, _ := json.Marshal(apostlePrice)
		_ = util.SetCache(ctx, key, bp, 60)
	} else {
		_ = json.Unmarshal(cache, &apostlePrice)
		for _, v := range FindOnsellApostle(ctx, "", chain, nil) {
			tokenMap[v.TokenId] = util.Evo.GetToken(GetChainByDistrict(v.District), v.Currency)
		}
		for _, v := range FindSiringApostlePrice(ctx, chain) {
			tokenMap[v.TokenId] = util.Evo.GetToken(GetChainByDistrict(v.District), v.Currency)
		}
		for _, v := range FindWorkerPrice(ctx, chain) {
			tokenMap[v.TokenId] = util.Evo.GetToken(GetChainByDistrict(v.District), v.Currency)
		}
	}
	return apostlePrice, tokenMap
}
