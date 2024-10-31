package models

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type Member struct {
	gorm.Model
	Name        string `gorm:"size:255;unique_index" json:"name"`
	Email       string `sql:"default: null" json:"email"`
	Wallet      string `sql:"default: null" json:"wallet"` // mysql 对字段内大小写不敏感
	TronWallet  string `sql:"default: null" json:"tron_wallet"`
	ItCode      string `gorm:"type:varchar(255);unique_index" json:"it_code"`
	Ancestry    string `gorm:"type:varchar(255)" json:"ancestry"`
	IteringId   uint   `json:"itering_id"`
	PlayerRole  int    `json:"player_role"`
	Mobile      string `json:"mobile" sql:"default: null;"`
	Region      string `json:"region"`
	Newbie      string `json:"newbie" sql:"default:'unfinished';"` // unfinished，finished，rewarded
	RewardChain string `json:"reward_chain"`
	OriTakeBackNonce
}

type MemberLoginInfo struct {
	gorm.Model
	MemberId uint   `json:"member_id"`
	Ip       string `json:"ip"`
	Ua       string `json:"ua"`
}

type ValidateLogin struct {
	Wallet    string `form:"wallet" json:"wallet" binding:"required"`
	Sign      string `form:"sign" json:"sign" binding:"required"`
	Chain     string `form:"chain" json:"chain"`
	Challenge string `form:"challenge" json:"challenge"`
}

type ValidateReg struct {
	Wallet    string `form:"wallet" json:"wallet" binding:"required"`
	Sign      string `form:"sign" json:"sign" binding:"required"`
	Chain     string `form:"chain" json:"chain"`
	Challenge string `form:"challenge" json:"challenge"`
	Name      string `form:"name" json:"name" binding:"required"`
	IteringId uint   `json:"itering_id"`
	ItCode    string `form:"itCode" json:"itCode"`
}

type MemberQueryField struct {
	Id         uint   `json:"id"`
	Wallet     string `form:"wallet" json:"wallet"`
	TronWallet string `form:"tron_wallet" json:"tron_wallet"`
	Email      string `form:"email" json:"email"`
	Name       string `form:"name" json:"name" binding:"required"`
	ItCode     string `form:"itCode" json:"itCode"`
	Mobile     string `form:"mobile" json:"mobile"`
	IteringId  uint   `json:"itering_id"`
}

type MemberJson struct {
	Id                 uint            `json:"id"`
	Email              string          `json:"email"`
	Wallet             string          `json:"wallet"`
	TronWallet         string          `json:"tron_wallet"`
	EthWallet          string          `json:"eth_wallet"`
	Name               string          `json:"name"`
	Mobile             string          `json:"mobile"`
	IsActive           int             `json:"is_active"`
	ItCode             string          `json:"itCode"`
	Balance            decimal.Decimal `json:"balance"`
	CooBalance         decimal.Decimal `json:"coo_balance"`
	KtonBalance        decimal.Decimal `json:"kton_balance"`
	RingWithdrawStatus bool            `json:"ring_withdraw_status"`
	IsInternal         bool            `json:"is_internal"` // 是否内测用户
	PlayerRole         int             `json:"player_role"`
	RegTime            int             `json:"reg_time"`
	Newbie             string          `json:"newbie"`
	RewardChain        string          `json:"reward_chain"`
}

func (m *Member) AfterCreate(tx *gorm.DB) (err error) {

	return m.fillAccountMemberId(&util.GormDB{DB: tx})
}

func (m *MemberLoginInfo) New(ctx context.Context) error {
	db := util.WithContextDb(ctx)
	var mli = new(MemberLoginInfo)
	if db.Where("member_id = ?", m.MemberId).First(&mli).RecordNotFound() {
		return db.Create(m).Error
	}
	if m.Ua == mli.Ua && m.Ip == mli.Ip {
		return nil
	}
	return db.Model(new(MemberLoginInfo)).Where("id = ?", mli.ID).Updates(map[string]interface{}{
		"ip": m.Ip,
		"ua": m.Ua,
	}).Error
}

func NewMember(ctx context.Context, chain, name, wallet string, iteringId uint) (uint, error) {
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	var member = Member{
		Name:      name,
		IteringId: iteringId,
	}
	member.fillAddress(wallet, chain)
	if query := db.Create(&member); query.Error != nil {
		return 0, query.Error
	}
	db.DbCommit()
	return member.ID, nil
}

func (m *Member) fillAddress(address, chain string) {
	switch chain {
	case EthChain:
		m.Wallet = address
	case TronChain:
		m.TronWallet = address
	default:
		m.Wallet = address
	}
}

func (m *Member) GetUseAddress(chain string) string {
	switch chain {
	case TronChain:
		return m.TronWallet
	default:
		return m.Wallet
	}
}

func (m *Member) SetUseAddress(chain, wallet string) *Member {
	switch chain {
	case TronChain:
		m.TronWallet = wallet
	default:
		m.Wallet = wallet
	}
	return m
}

func (m *Member) GetUseWithdrawNonce(ctx context.Context, chain, currency string) int {
	ethNonce := map[string]int{currencyRing: m.WithdrawNonce, currencyKton: m.KtonNonce}
	tronNonce := map[string]int{currencyRing: m.TronNonce, currencyKton: m.TronKtonNonce}
	switch chain {
	case TronChain:
		return tronNonce[currency]
	case EthChain:
		return ethNonce[currency]
	default:
		takeBack := m.TakeBackNonce(ctx)
		if takeBack == nil {
			return 0
		}
		nonceField := strings.ToLower(fmt.Sprintf("%s_%s", chain, currency))
		if val, _ := util.GetFieldValByTag(nonceField, "json", takeBack); val != nil {
			return val.(int)
		}
		return 0
	}
}

func (m *Member) fillAccountMemberId(db *util.GormDB) error {
	var account Account
	if m.Wallet != "" {
		query := db.Where("wallet = ?", m.Wallet).Find(&account)
		if !query.RecordNotFound() {
			if err := db.Model(&account).Where("wallet = ?", m.Wallet).UpdateColumns(Account{MemberId: m.ID}).Error; err != nil {
				return err
			}
		}
	}
	if m.TronWallet != "" {
		query := db.Where("wallet = ?", m.TronWallet).Find(&account)
		if !query.RecordNotFound() {
			if err := db.Model(&account).Where("wallet = ?", m.TronWallet).UpdateColumns(Account{MemberId: m.ID}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Member) GetWithdrawStatus(ctx context.Context, chain, currency string) bool {
	contract := util.GetContractAddress("takeBack", chain)
	if currency == currencyKton {
		contract = util.GetContractAddress("takeBackKton", chain)
	}
	sg := storage.New(chain)
	chainNonce := int(sg.UserToNonce(m.GetUseAddress(chain), contract))

	switch chain {
	case EthChain:
		if currency == currencyRing {
			return m.WithdrawNonce == chainNonce
		}
		return m.KtonNonce == chainNonce
	case TronChain:
		if currency == currencyRing {
			return m.TronNonce == chainNonce
		}
		return m.TronKtonNonce == chainNonce
	default:
		takeBack := m.TakeBackNonce(ctx)
		if takeBack == nil {
			return true
		}
		nonceField := strings.ToLower(fmt.Sprintf("%s_%s", chain, currency))
		if val, _ := util.GetFieldValByTag(nonceField, "json", takeBack); val != nil {
			return val.(int) == chainNonce
		}
		return true
	}
}

func (m *Member) updateField(ctx context.Context, db *util.GormDB) error {
	query := db.Model(&m).UpdateColumn(m)
	cacheKey := fmt.Sprintf("GetMemberByid:%d", m.ID)
	util.DelCache(ctx, cacheKey)
	return query.Error
}

func (m *Member) BindEmail(ctx context.Context, email string) {
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	m.Email = email
	_ = m.updateField(ctx, db)
	db.DbCommit()
}

func (m *Member) TouchAccount(ctx context.Context, currency, wallet string, chain string, noMemberId ...bool) *Account {
	db := util.WithContextDb(ctx)
	account := Account{}
	if wallet == "" {
		return &account
	}
	query := map[string]interface{}{
		"currency": currency,
		"wallet":   wallet,
		"chain":    chain,
	}
	if len(noMemberId) == 0 || !noMemberId[0] {
		query["member_id"] = m.ID
	}
	db.FirstOrCreate(&account, query)
	return &account
}

func (m *Member) GetBalance(ctx context.Context, code, wallet, chain string) decimal.Decimal {
	return m.TouchAccount(ctx, code, wallet, chain).Balance
}

func (m *Member) updateWithdrawNonce(ctx context.Context, db *util.GormDB, nonce int, chain, currency string) (err error) {
	field := "withdraw_nonce"

	var updateNonce = func() error {
		query := db.Model(&m).UpdateColumn(map[string]interface{}{field: nonce})
		util.DelCache(ctx, fmt.Sprintf("GetMemberByid:%d", m.ID))
		return query.Error
	}
	switch chain {
	case TronChain:
		field = map[string]string{currencyRing: "tron_nonce", currencyKton: "tron_kton_nonce"}[currency]
		return updateNonce()
	case EthChain:
		field = map[string]string{currencyRing: "withdraw_nonce", currencyKton: "kton_nonce"}[currency]
		return updateNonce()
	default:
		field = fmt.Sprintf("%s_%s", chain, currency)
		db.Model(MemberTakeBack{}).Where("member_id = ?", m.ID).
			UpdateColumn(map[string]interface{}{field: nonce})
	}
	return
}

func (m *Member) UploadPlayRoleByKey(ctx context.Context, db *util.GormDB, key string) bool {
	if m.PlayerRole != 0 {
		return false
	}
	m.PlayerRole = 1
	if useKeyStore(ctx, m.ID, key) {
		_ = m.updateField(ctx, db)
	}
	return false
}

func (m *MemberQueryField) GetMemberBy(ctx context.Context, field string) *Member {
	db := util.WithContextDb(ctx)
	member := Member{}
	value, boolErr := getStringValueByFieldName(m, field)
	if !boolErr {
		return nil
	}
	query := db.Where(util.SnakeString(field)+" = ?", value).First(&member)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	return &member
}

func AuthOwner(c *gin.Context, force ...bool) *Member {
	if owner := c.Query("owner"); owner != "" {
		if member := GetMemberByAddress(util.GetContextByGin(c), owner, c.GetString("EvoNetwork")); member != nil {
			return member
		}
		if len(force) != 0 && force[0] {
			return new(Member).SetUseAddress(c.GetString("EvoNetwork"), owner)
		}
	}
	return nil
}

func GetMember(ctx context.Context, id int) *Member {
	db := util.WithContextDb(ctx)
	var member Member
	if id == 0 {
		return nil
	}
	cacheKey := fmt.Sprintf("GetMemberByid:%d", id)
	if cache := util.GetCache(ctx, cacheKey); cache != nil {
		_ = json.Unmarshal(cache, &member)
	} else {
		query := db.First(&member, id)
		if query.Error != nil || query == nil || query.RecordNotFound() {
			return nil
		}
		cache, _ = json.Marshal(member)
		_ = util.SetCache(ctx, cacheKey, cache, 3600)
	}
	return &member
}

// chain is require, can fill chain type like Eth,Tron,PChain
func GetMemberByAddress(ctx context.Context, address string, chain string) *Member {
	if address == "" {
		return nil
	}
	query := MemberQueryField{Wallet: address, TronWallet: address}
	switch chain {
	case TronChain:
		return query.GetMemberBy(ctx, "TronWallet")
	default:
		return query.GetMemberBy(ctx, "Wallet")
	}
}

func GetOrCreateMemberByAddress(ctx context.Context, address string, chain string) *Member {
	if address == "" {
		return nil
	}
	m := GetMemberByAddress(ctx, address, chain)
	if m != nil {
		return nil
	}
	_, _ = NewMember(ctx, chain, address, address, 0)
	return GetMemberByAddress(ctx, address, chain)
}

func (m *Member) RefreshMemberLoginTime(ctx context.Context) {
	util.WithContextDb(ctx).Model(&m).Update("updated_at", time.Now())
}

func GetMemberCount(ctx context.Context, chain string) uint64 {
	var count uint64
	switch chain {
	case TronChain:
		util.WithContextDb(ctx).Table("members").Where("tron_wallet is not null").Count(&count)
	default:
		util.WithContextDb(ctx).Table("members").Where("wallet is not null").Count(&count)
	}
	return count
}

func GetMemberActive24H(ctx context.Context, chain string) uint64 {
	var count uint64
	switch chain {
	case TronChain:
		util.WithContextDb(ctx).Table("members").Where("tron_wallet is not null").Where("updated_at >= ?", time.Now().UTC().Unix()-86400).Count(&count)
	default:
		util.WithContextDb(ctx).Table("members").Where("wallet is not null").Where("updated_at >= ?", time.Now().UTC().Unix()-86400).Count(&count)
	}
	return count
}

func GetMemberActive7D(ctx context.Context, chain string) uint64 {
	var count uint64
	switch chain {
	case TronChain:
		util.WithContextDb(ctx).Table("members").Where("tron_wallet is not null").Where("updated_at >= ?", time.Now().UTC().Unix()-7*86400).Count(&count)
	default:
		util.WithContextDb(ctx).Table("members").Where("wallet is not null").Where("updated_at >= ?", time.Now().UTC().Unix()-7*86400).Count(&count)
	}
	return count
}
