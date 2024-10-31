package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

type Land struct {
	gorm.Model
	Owner        string `json:"owner"`
	Status       string `json:"status"`
	Lon          int    `json:"lon"` // 经度
	Lat          int    `json:"lat"` // 纬度
	GX           int    `json:"gx"`  // 游戏内经度
	GY           int    `json:"gy"`  // 游戏内纬度
	MemberId     int    `json:"member_id"`
	District     int    `json:"district"`
	Introduction string `json:"introduction" sql:"type:text;"`
	TokenId      string `json:"token_id"`
	Chain        string `json:"chain"`
	TokenIndex   int    `json:"token_index"`
	LocationId   string `json:"location_id"`
	Cover        string `json:"cover"`
	LandUrl      string `json:"land_url"`
	TransTime    int    `json:"trans_time"`
	DappId       uint   `sql:"default: 0" json:"dapp_id"`
	Sticker      string `json:"sticker"`
	BuildingId   uint   `json:"building_id" sql:"default:0"`
}

type LandJson struct {
	ID      uint   `json:"id" example:"1"`
	Owner   string `json:"owner" example:"0x9273283412f0A26C2cB99BBD874D54AD18540101"`
	Status  string `json:"status" example:"fresh" enums:"[fresh,onsell]"`
	TokenId string `json:"token_id" example:"2a0100010100010100000000000000010000000000000000000000000000000c"`
	// longitude
	Lon int `json:"lon" example:"-76"`
	// latitude
	Lat int `json:"lat" example:"-2"`
	// longitude in game map
	GX int `json:"gx" example:"31"`
	// latitude in game map
	GY int `json:"gy" example:"12"`
	// only setting by owner
	Introduction string `json:"introduction" sql:"type:text;" example:"Hi"`
	// member id
	MemberId int `json:"member_id"`

	// land name
	Name string `json:"name"`
	// last bid is mine
	MineLastBid bool `json:"mine_last_bid"`
	// resource on land
	Resource  LandDataJson `json:"resource"`
	PendingTx string       `json:"pending_tx"`
	// Has somebody bid on land
	HasBid bool  `json:"has_bid"`
	LandId int64 `json:"land_id"`
	// dapp picture on land
	Cover string `json:"cover"`
	// if cover if empty, must be use it
	Picture string `json:"picture"`
	// token index
	TokenIndex int `json:"token_index"`
	// dapp picture on land
	LandUrl string `json:"land_url"`
	// apostles who work on land
	ApostleWorker  []ApostleMiner  `json:"apostle_worker"`
	CurrentPrice   decimal.Decimal `json:"current_price"`
	Sticker        string          `json:"sticker"`
	Drills         []LandEquip     `json:"drills"` // Drill on land
	AuctionStartAt int             `json:"auction_start_at"`
	District       int             `json:"district"`
	Token          *util.Token     `json:"token"`
}

type LandJoinLandData struct {
	Owner   string `json:"owner"`
	Status  string `json:"status"`
	TokenId string `json:"token_id"`
	Lon     int    `json:"lon"` // 经度
	Lat     int    `json:"lat"` // 纬度
	GX      int    `json:"gx"`  // 游戏内经度
	GY      int    `json:"gy"`  // 游戏内纬度
	Cover   string `json:"cover"`
	Sticker string `json:"sticker"`
	LandDataJson
}

type SampleLand struct {
	District      int             `json:"district"`
	TokenId       string          `json:"token_id"`
	Owner         string          `json:"owner"`
	Status        string          `json:"status"`
	Lon           int             `json:"lon"` // 经度
	Lat           int             `json:"lat"` // 纬度
	GX            int             `json:"gx"`  // 游戏内经度
	GY            int             `json:"gy"`  // 游戏内纬度
	Resource      string          `json:"resource"`
	AuctionStatus string          `json:"as"`
	Genesis       bool            `json:"genesis"`
	Cover         string          `json:"cover"`
	Picture       string          `json:"picture"`
	LandId        int64           `json:"land_id"`
	CurrentPrice  decimal.Decimal `json:"current_price"`
	Sticker       string          `json:"sticker"`
	BuildingId    int             `json:"building_id"`
	LandDataJson  `json:"-"`
	Token         *util.Token `json:"token"`
}

type LandDetailJson struct {
	LandData          LandJson                 `json:"land_data"`
	Auction           *AuctionJson             `json:"auction"` // auction data
	Record            []LandAuctionHistoryJson `json:"record"`
	Resource          LandDataJson             `json:"resource"`
	LandAttenuationAt int                      `json:"land_attenuation_at"`
	Dapp              *DappJson                `json:"dapp"` // dapp data
}

type ValidateIntroduction struct {
	Introduction string `form:"introduction" json:"introduction"`
	TokenId      string `form:"token_id" json:"token_id" binding:"required"`
	Cover        string `form:"cover" json:"cover"`
	LandUrl      string `form:"land_url" json:"land_url"`
}

type LandQuery struct {
	WhereQuery struct {
		Owner    string `json:"owner" table_name:"lands"`
		District int    `json:"district" table_name:"lands"`
		Status   string `json:"status" table_name:"lands"`
		TokenId  string `json:"token_id" table_name:"lands"`
	}
	MultiFilter struct {
		Element []string
		Price   map[string]string
		Flag    []string
		Owner   []string
	}
	WhereInterface []string
	TokenId        []string
	Filter         string
	MyLastBid      []string
	PendingTrans   map[string]string
	HasBid         []string
	TokenMap       map[string]*util.Token
	JoinGoldRush   []string
	Network        string
	Row            int
	Page           int
	OrderField     string
	Order          string
	PriceMap       map[string]decimal.Decimal

	AuctionStartAtMap map[string]int
}

var landOrder = []string{"price", "gold_rate", "wood_rate", "water_rate", "fire_rate", "soil_rate"}

func GenerateLandTokenId(chain string, tokenIndex int) string {
	locationIndex := fmt.Sprintf("%032s", fmt.Sprintf("%x", tokenIndex))
	tokenIndexPrefix := "2a010001010001010000000000000001%s"
	if prefix, ok := util.Evo.TokenIdPrefix[chain]; ok {
		tokenIndexPrefix = prefix + "%s"
	}
	return fmt.Sprintf(tokenIndexPrefix, locationIndex)
}

func (l *Land) New(db *util.GormDB) error {
	var land = Land{
		Owner:      l.Owner,
		Status:     landFresh,
		Lon:        l.Lon,
		Lat:        l.Lat,
		MemberId:   l.MemberId,
		TokenId:    l.TokenId,
		District:   l.District,
		Chain:      l.Chain,
		TokenIndex: l.TokenIndex,
	}
	result := db.Create(&land)
	return result.Error
}

func (l *Land) AfterCreate(tx *gorm.DB) {
	ld, _ := GetLandData(l.TokenId)
	ld.LandId = l.ID
	_ = ld.New(context.TODO(), tx)
}

func (l *Land) AsJson(ctx context.Context, memberInfo *Member) *LandDetailJson {
	if l == nil {
		return nil
	}
	chain := GetChainByTokenId(l.TokenId)
	landId := new(big.Int).Mod(util.U256(l.TokenId), big.NewInt(65536)).Int64()
	lj := LandDetailJson{LandData: LandJson{
		Owner: l.Owner, Status: l.Status, Lat: l.Lat, Lon: l.Lon, TokenId: l.TokenId, Introduction: l.Introduction, MemberId: l.MemberId,
		LandId: landId, Cover: l.Cover, LandUrl: l.LandUrl, ID: l.ID,
		GX: l.GX, GY: l.GY, Sticker: l.Sticker, District: l.District,
	}}
	lj.Dapp = l.Dapp(ctx, memberInfo)
	lj.LandAttenuationAt = l.getLandAttenuationAt()
	member := GetMemberByAddress(ctx, l.Owner, chain)
	if member != nil {
		lj.LandData.Name = member.Name
	}
	auction := GetCurrentAuction(ctx, l.TokenId, AuctionGoing)
	if auction != nil {
		if auction.StartAt > int(time.Now().Unix()) {
			lj.LandData.Status = landFresh
		}
		lj.Auction = auction.AsJson(ctx)
	}
	lj.LandData.ApostleWorker = lj.LandData.apostles(ctx)
	lj.Record = landAuctionHistory(ctx, l.TokenId)
	lj.Resource = landResources(ctx, l.TokenId)
	lj.LandData.Resource = lj.Resource
	// drill
	lj.LandData.Drills = l.drills(ctx)
	if lj.Auction != nil {
		lj.LandData.Token = lj.Auction.Token
	}
	lj.LandData.Picture = GetLandPicture(l.District, l.GX, l.GY)
	return &lj
}

func (l *Land) getLandAttenuationAt() int {
	chain := GetChainByTokenId(l.TokenId)
	if chain == TronChain {
		return tronLandAttenuationAt
	}
	return ethLandAttenuationAt
}

func (l *Land) UnclaimedResource() map[string]decimal.Decimal {
	zero := decimal.Zero
	resources := map[string]decimal.Decimal{"gold": zero, "wood": zero, "water": zero, "fire": zero, "soil": zero}
	chain := GetChainByTokenId(l.TokenId)
	resourceAddress := []string{
		util.GetContractAddress("gold", chain),
		util.GetContractAddress("wood", chain),
		util.GetContractAddress("water", chain),
		util.GetContractAddress("fire", chain),
		util.GetContractAddress("soil", chain),
	}
	sg := storage.New(chain)
	result := sg.UnclaimedResource(l.TokenId, resourceAddress)
	if len(result) >= 5 {
		resources["gold"] = util.BigToDecimal(util.U256(result[0]))
		resources["wood"] = util.BigToDecimal(util.U256(result[1]))
		resources["water"] = util.BigToDecimal(util.U256(result[2]))
		resources["fire"] = util.BigToDecimal(util.U256(result[3]))
		resources["soil"] = util.BigToDecimal(util.U256(result[4]))
	}
	return resources
}

func (lq *LandQuery) LandList(ctx context.Context) (*[]LandJson, int) {
	db := util.WithContextDb(ctx)
	var land []LandJson
	var count int
	var ori *gorm.DB

	// where
	if lq.Filter != "" && !util.StringInSlice(lq.Filter, []string{"my", "fresh", "mine", "availableDrill", "other", "plo"}) {
		ori = db.Table("lands").Where(lq.WhereQuery).Where("token_id in (?)", lq.TokenId)
	} else if lq.Filter == "mine" {
		ori = db.Table("lands").Where(lq.WhereQuery).Or("token_id in (?)", lq.TokenId)
	} else {
		ori = db.Table("lands").Where(lq.WhereQuery)
	}
	for _, q := range lq.WhereInterface {
		ori = ori.Where(q)
	}
	// order
	if lq.OrderField == "" || lq.OrderField == "price" {
		ori = ori.Order("trans_time desc")
	} else {
		ori = ori.Order(fmt.Sprintf("%s %s", "token_index", lq.Order))
	}
	query := ori.Scan(&land)

	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil, 0
	}

	// pending
	var pendingLand []string
	if len(lq.PendingTrans) > 0 {
		for k := range lq.PendingTrans {
			pendingLand = append(pendingLand, k)
		}
	}

	for i, v := range land {
		land[i].Resource = landResources(ctx, v.TokenId)
		if land[i].Status == "onsell" { // n*1
			land[i].CurrentPrice = lq.PriceMap[v.TokenId]
			land[i].AuctionStartAt = lq.AuctionStartAtMap[v.TokenId]
		}
	}

	// order
	land = lq.sortLandJson(land)
	// filter
	land = lq.filterLandJson(land)
	// count
	count = len(land)
	if count == 0 {
		return nil, 0
	}
	// limit
	maxLength := (lq.Page + 1) * lq.Row
	start := lq.Page * lq.Row
	if maxLength > count {
		maxLength = count
	}
	if start > count {
		start = count - maxLength
	}

	land = land[start:maxLength]
	if lq.TokenMap == nil {
		lq.TokenMap = make(map[string]*util.Token)
	}

	// render
	for i, v := range land {
		land[i].LandId = int64(v.TokenIndex + 1)
		if util.StringInSlice(v.TokenId, lq.MyLastBid) {
			land[i].MineLastBid = true
		}
		if util.StringInSlice(v.TokenId, pendingLand) {
			land[i].PendingTx = lq.PendingTrans[v.TokenId]
		}
		if util.StringInSlice(v.TokenId, lq.HasBid) {
			land[i].HasBid = true
		}
		land[i].ApostleWorker = v.apostles(ctx)
		land[i].Cover = v.Cover

		landModel := Land{TokenId: v.TokenId}
		land[i].Drills = landModel.drills(ctx)
		if _, ok := lq.TokenMap[land[i].TokenId]; ok {
			land[i].Token = lq.TokenMap[land[i].TokenId]
		}
		land[i].Picture = GetLandPicture(v.District, v.GX, v.GY)
	}
	return &land, count
}

func GetLandPicture(district, gx, gy int) string {
	districtMap := map[int]string{1: "a", 2: "b", 3: "c", 4: "d", 5: "e"}
	// districtMap[land.District]
	districtStr := districtMap[district]
	if districtStr == "" {
		districtStr = cast.ToString(district)
	}
	return fmt.Sprintf("https://gcs.evolution.land/land/%s/g%d-%d.png", districtStr, gx, gy)
}

func (lq *LandQuery) AllLands(ctx context.Context) (lands []SampleLand, count int) {
	db := util.WithContextDb(ctx).Table("lands")
	wheres, values := util.StructToSql(lq.WhereQuery)
	if len(wheres) > 0 {
		db = db.Where(strings.Join(wheres, " AND "), values...)
	}
	ori := db.
		Select("owner,status,lands.token_id,lon,lat,is_reserved,gold_rate," +
			"wood_rate,water_rate,fire_rate,soil_rate,is_special,has_box,cover,gx,gy,sticker,building_id," +
			"district").
		Joins("left join land_data on lands.id = land_data.land_id")
	if len(lq.TokenId) > 0 {
		ori = ori.Where("lands.token_id in (?)", lq.TokenId)
	}
	for _, q := range lq.WhereInterface {
		ori = ori.Where(fmt.Sprintf("lands.%s", q))
	}

	if lq.OrderField != "price" {
		ori = ori.Order(fmt.Sprintf("%s %s", lq.OrderField, lq.Order))
	}
	query := ori.Scan(&lands)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil, 0
	}

	count = len(lands)

	tokenIds := MyAuctionLandList(ctx, lq.WhereQuery.District, nil)
	notStartTokenIds := notStartLands(ctx)
	genesisTokenIds := genesisLands(ctx, lq.Network)
	if lq.TokenMap == nil {
		lq.TokenMap = make(map[string]*util.Token)
	}

	for index, land := range lands {
		resource := fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d", land.IsReserved, land.GoldRate, land.WoodRate, land.WaterRate, land.FireRate, land.SoilRate, land.IsSpecial, land.HasBox)
		if land.HasBox == 1 {
			resource = "0,0,0,0,0,0,0,1"
		}
		lands[index].Resource = resource
		if util.StringInSlice(land.TokenId, tokenIds) {
			lands[index].AuctionStatus = AuctionClaimed
		}
		if util.StringInSlice(land.TokenId, notStartTokenIds) {
			lands[index].Status = landFresh
		}
		if util.StringInSlice(land.TokenId, genesisTokenIds) {
			lands[index].Genesis = true
		}
		if land.Status == landOnsell && land.AuctionStatus == "" {
			lands[index].CurrentPrice = lq.PriceMap[land.TokenId]
		}
		lands[index].Cover = land.Cover
		lands[index].LandId = new(big.Int).Mod(util.U256(land.TokenId), big.NewInt(65536)).Int64()
		if _, ok := lq.TokenMap[lands[index].TokenId]; ok {
			lands[index].Token = lq.TokenMap[lands[index].TokenId]
		}
		lands[index].Picture = GetLandPicture(land.District, land.GX, land.GY)
	}

	lands = lq.sortSampleLands(lands)
	maxLength := (lq.Page + 1) * lq.Row
	if maxLength > count {
		maxLength = count
	}
	lands = lands[lq.Page*lq.Row : maxLength]
	return
}

func (lq *LandQuery) sortSampleLands(lands []SampleLand) []SampleLand {
	if lq.OrderField == "price" {
		sort.Slice(lands[:], func(i, j int) bool {
			if lq.Order == "desc" {
				i, j = j, i
			}
			return lands[i].CurrentPrice.LessThan(lands[j].CurrentPrice)
		})
	}
	return lands
}

func (lq *LandQuery) sortLandJson(lands []LandJson) []LandJson {
	maxValue := decimal.New(1, util.GetTokenDecimals(lq.Network))
	if util.StringInSlice(lq.OrderField, landOrder) {
		sort.Slice(lands[:], func(i, j int) bool {
			if lq.Order == "desc" {
				i, j = j, i
			}
			switch lq.OrderField {
			case "price":
				prev := lands[i].CurrentPrice
				if !prev.IsPositive() {
					prev = maxValue
				}
				next := lands[j].CurrentPrice
				if !next.IsPositive() {
					next = maxValue
				}
				return prev.LessThan(next)
			default:
				valuePrev, _ := getStringValueByFieldName(lands[i].Resource, util.CamelString(lq.OrderField))
				valueNext, _ := getStringValueByFieldName(lands[j].Resource, util.CamelString(lq.OrderField))
				return util.StringToInt(valuePrev) < util.StringToInt(valueNext)
			}
		})
	}
	return lands
}

// query filter
// element=gold&element=wood&element=hoo&element=fire&element=soil
// price[gte]=1&price[lte]=10
// flag=normal&flag=reserved&flag=box
// owner=mine&&owner=other

type PriceCompare struct {
	Gte *decimal.Decimal `json:"gte,omitempty"`
	Lte *decimal.Decimal `json:"lte,omitempty"`
}

func (lq *LandQuery) filterLandJson(lands []LandJson) []LandJson {
	multiFilter := lq.MultiFilter
	var filterLand []LandJson
	var compare PriceCompare

	elementMap := map[string]int{currencyGold: 0, currencyWood: 1, currencyWater: 2, currencyFire: 3, currencySoil: 4}

	for _, land := range lands {

		// element
		if len(multiFilter.Element) > 0 && len(multiFilter.Element) < 5 {
			resources := []int{land.Resource.GoldRate, land.Resource.WoodRate, land.Resource.WoodRate, land.Resource.FireRate, land.Resource.SoilRate}
			index := util.IntsMaxIndex(resources)
			var pass bool
			for _, element := range multiFilter.Element {
				if util.IntInSlice(elementMap[element], index) {
					pass = true
					break
				}
			}
			if !pass {
				continue
			}
		}

		if len(multiFilter.Price) > 0 {
			util.UnmarshalAny(&compare, multiFilter.Price)
			if compare.Gte != nil && land.CurrentPrice.LessThan(*compare.Gte) {
				continue
			}
			if compare.Lte != nil && land.CurrentPrice.GreaterThan(*compare.Lte) {
				continue
			}
		}
		if len(multiFilter.Flag) > 0 && len(multiFilter.Flag) < 3 {
			var pass bool
			for _, flag := range multiFilter.Flag {
				if flag == "normal" && land.Resource.HasBox == 0 && land.Resource.IsReserved == 0 {
					pass = true
					break
				}
				if flag == "reserved" && land.Resource.IsReserved == 1 {
					pass = true
					break
				}
				if flag == "box" && land.Resource.HasBox == 1 {
					pass = true
					break
				}
			}
			if !pass {
				continue
			}
		}
		filterLand = append(filterLand, land)
	}

	return filterLand
}

func (vi *ValidateIntroduction) UpdateIntroduction(ctx context.Context, memberId uint) error {
	db := util.WithContextDb(ctx)
	var land Land
	query := db.Where("token_id = ? and member_id = ?", vi.TokenId, memberId).First(&land)
	if !query.RecordNotFound() {
		db.Model(&land).Where("token_id = ? and member_id = ?", vi.TokenId, memberId).
			UpdateColumn(map[string]interface{}{"introduction": vi.Introduction, "cover": vi.Cover, "land_url": vi.LandUrl})
		return nil
	} else {
		return errors.New("no permission")
	}
}

func GetLandByTokenId(ctx context.Context, tokenId string) *Land {
	db := util.WithContextDb(ctx)
	var land Land
	query := db.Table("lands").Where("token_id = ?", tokenId).Scan(&land)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}

	return &land
}

func GetLandByCoord(ctx context.Context, x, y int) *Land {
	db := util.WithContextDb(ctx)
	var land Land
	query := db.Table("lands").Where("lon = ? and lat=?", x, y).Scan(&land)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	return &land
}

func InitLandsFormChain(ctx context.Context, chain string) {
	count := 2025
	for tokenIndex := 0; tokenIndex < count; tokenIndex++ {
		tokenId := GenerateLandTokenId(chain, tokenIndex+1)
		if GetLandByTokenId(ctx, tokenId) != nil {
			continue
		}
		if err := fillLand(ctx, tokenIndex, tokenId); err != nil {
			log.Debug("fill land error. tokenIndex=%d. error: %s", tokenIndex, err)
			break
		}
	}
	log.Debug("Success fill lands")
}

func RefreshLands(ctx context.Context, chain string) {
	count := 2025
	for tokenIndex := 0; tokenIndex < count; tokenIndex++ {
		tokenId := GenerateLandTokenId(chain, tokenIndex+1)
		_ = refreshLandData(ctx, tokenId)
	}
	log.Debug("Success fill lands")
}

func fillLand(ctx context.Context, tokenIndex int, tokenId string) error {
	db := util.DbBegin(ctx)
	if err := createByTokenId(db, tokenId, tokenIndex); err != nil {
		db.Rollback()
		return err
	}
	db.DbCommit()
	return nil
}

func createByTokenId(db *util.GormDB, tokenId string, tokenIndex int) error {
	chain := GetChainByTokenId(tokenId)
	sg := storage.New(chain)
	lon, lat, err := sg.GetTokenLocationHM(tokenId)
	if err != nil {
		return err
	}

	owner, _ := sg.OwnerOf(tokenId)
	if owner == "" {
		return errors.New("get land owner fail")
	}
	land := Land{Status: landFresh, Chain: "ethereum", Owner: owner, TokenIndex: tokenIndex, TokenId: util.TrimHex(tokenId), Lon: int(lon), Lat: int(lat), MemberId: 0, District: getNFTDistrict(tokenId)}
	if chain != EthChain {
		land.Chain = chain
	}
	return land.New(db)
}

func updateLandOwner(ctx context.Context, db *util.GormDB, tokenId, owner string) error {
	var land Land
	query := db.Where("token_id = ?", tokenId).First(&land)
	if !query.RecordNotFound() {
		if member := GetMemberByAddress(ctx, owner, GetChainByTokenId(tokenId)); member != nil {
			memberId := member.ID
			query = db.Model(&land).Where("token_id = ?", tokenId).UpdateColumn(Land{Owner: owner, MemberId: int(memberId), TransTime: int(time.Now().Unix())})
		} else {
			query = db.Model(&land).Where("token_id = ?", tokenId).UpdateColumn(Land{Owner: owner, TransTime: int(time.Now().Unix())})
		}
	}
	return query.Error
}

func getLand(ctx context.Context, tokenId string) *Land {
	db := util.WithContextDb(ctx)
	var land Land
	query := db.Where("token_id = ?", tokenId).Find(&land)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	return &land
}

func (l *Land) landName() string {
	symbol := "A"
	landId := new(big.Int).Mod(util.U256(l.TokenId), big.NewInt(65536)).Int64()
	if l.District == 2 {
		symbol = "B"
	}
	return fmt.Sprintf("%s%d", symbol, landId)
}

func (l *Land) DelDapp(ctx context.Context, dapp *Dapp) error {
	txn := util.DbBegin(ctx)
	defer txn.DbRollback()
	query := txn.Model(&l).UpdateColumn(map[string]interface{}{"dapp_id": 0})
	if query.Error != nil {
		txn.DbRollback()
		return errors.New("Del error")
	}
	if err := dapp.Del(txn); err != nil {
		txn.DbRollback()
		return errors.New("Del error")
	}
	txn.DbCommit()
	return nil
}

func (l *Land) Dapp(ctx context.Context, memberInfo *Member) *DappJson {
	if l.DappId <= 0 && memberInfo != nil && int(memberInfo.ID) != l.MemberId {
		return nil
	}
	db := util.WithContextDb(ctx)
	var dapp DappJson
	var query *gorm.DB
	if l.DappId > 0 && (memberInfo == nil || int(memberInfo.ID) != l.MemberId) {
		query = db.Table("dapps").Where("id = ?", l.DappId).Scan(&dapp)
	} else {
		query = db.Table("dapps").Where("land_id=?", l.ID).Order("id asc").Scan(&dapp)
	}
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	return &dapp
}

func AddLandsGameXY(ctx context.Context, district int) {
	b, err := os.ReadFile(fmt.Sprintf("data/coordinate%d.json", district))
	util.Panic(err)
	var coordMap map[string]string
	_ = json.Unmarshal(b, &coordMap)
	var lands []Land
	db := util.WithContextDb(ctx)
	db.Table("lands").Where("district = ?", district).Find(&lands)
	for _, land := range lands {
		coord, ok := coordMap[fmt.Sprintf("%d|%d", land.Lon, land.Lat)]
		if ok {
			gameCoord := strings.Split(coord, "|")
			db.Model(Land{}).Where("lon=? and lat=?", land.Lon, land.Lat).UpdateColumn(map[string]interface{}{
				"gx": util.StringToInt(gameCoord[0]),
				"gy": util.StringToInt(gameCoord[1]),
			})
		}
	}
	log.Debug("Success")
}

// 判断land 是否相邻，建筑地块选择选择左下角的地块
// 地块分为 1*1, 2*2, 3*3

func checkLandNear(lon, lat, drawingIndex int) []string {
	startX, startY := lon, lat
	var list []string
	landRange := drawingIndex * drawingIndex
	for i := 0; i < landRange; i++ {
		if i%drawingIndex == 0 {
			list = append(list, fmt.Sprintf("%d,%d", startX, startY+i/drawingIndex))
			continue
		}
		list = append(list, fmt.Sprintf("%d,%d", startX+i%drawingIndex, startY+i/drawingIndex))
	}
	return list
}

func setLandBuilding(txn *util.GormDB, buildId uint, tokenIds []string) error {
	query := txn.Model(Land{}).Where("building_id = 0").Where("token_id in (?)", tokenIds).UpdateColumn(Land{BuildingId: buildId})
	if int(query.RowsAffected) != len(tokenIds) {
		return errors.New("land building update fail")
	}
	return nil
}

func (l *Land) Resource(ctx context.Context) LandDataJson {
	return landResources(ctx, l.TokenId)
}

func (l *Land) drills(ctx context.Context) []LandEquip {
	var list []LandEquip
	util.WithContextDb(ctx).Model(LandEquip{}).Where("land_token_id = ?", l.TokenId).Order("`index` asc").Find(&list)
	formulas := util.Evo.Formula[GetChainByTokenId(l.TokenId)]
	formulasMap := make(map[int]util.Formula)
	for _, formula := range formulas {
		formulasMap[formula.Id] = formula
	}
	for index, v := range list {
		class := formulasMap[v.FormulaId].Class
		list[index].EquipTime = drillProtectPeriod(ctx, v.DrillTokenId, class, v.EquipTime)
	}
	return list
}

type LandRank struct {
	Count int    `json:"count"` // number of lands
	Owner string `json:"owner"` // owner address
	Name  string `json:"name"`
}

func LandsRankList(ctx context.Context, district int, chain string) []LandRank {
	type GroupLandOwner struct {
		Count int
		Owner string
	}
	var groups []GroupLandOwner
	var ranks []LandRank
	db := util.WithContextDb(ctx).Model(Land{})
	db = db.Select("count(*) as count, owner")
	db = db.Where("owner not in (?)",
		[]string{
			util.GetContractAddress("clockAuction", chain),
			util.GetContractAddress("genesisHolder", chain),
			util.GetContractAddress("claims", chain),
		})
	db.Where("district = ?", district).Group("owner").Limit(10).Order("count desc").Scan(&groups)

	for _, v := range groups {
		rank := LandRank{Count: v.Count, Owner: v.Owner}
		if member := GetMemberByAddress(ctx, v.Owner, chain); member != nil {
			rank.Name = member.Name
		}
		ranks = append(ranks, rank)
	}
	return ranks
}
