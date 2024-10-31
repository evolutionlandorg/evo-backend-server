package models

import (
	"context"
	"errors"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"

	solsha3 "github.com/evolutionlandorg/evo-backend/pkg/github.com/miguelmota/go-solidity-sha3"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"
)

type Treasure struct {
	gorm.Model
	Status    string          `json:"status"` // locked,unlock,used
	KeyStore  string          `json:"key_store"`
	MemberId  uint            `json:"member_id"`
	TxId      string          `json:"tx_id"`
	BoxType   string          `json:"box_type"`  // gold,silver
	BoxIndex  string          `json:"box_index"` // uuid
	Buyer     string          `json:"buyer"`
	RingValue decimal.Decimal `json:"ring_value" sql:"type:decimal(32,16);"`
	CooValue  decimal.Decimal `json:"coo_value" sql:"type:decimal(32,16);"`
	BuyTime   int             `json:"buy_time"`
	Chain     string          `json:"chain"`
	Gen       int             `json:"gen"`
	TokenId   string          `json:"token_id"`
	Price     decimal.Decimal `json:"price" sql:"type:decimal(32,0);"`
	OpenTime  int             `json:"open_time"`
	IsLocked  bool            `json:"-" sql:"default:0"`
}

type TreasureQuery struct {
	WhereQuery struct {
		MemberId uint
		Buyer    string
		Status   string
		Gen      int
		Chain    string
	}
	Row  int
	Page int
}

type TreasureJson struct {
	BoxType      string          `json:"box_type"` // gold,silver
	RingValue    decimal.Decimal `json:"ring_value"`
	CooValue     decimal.Decimal `json:"coo_value"`
	BoxIndex     string          `json:"box_index"`
	Status       string          `json:"status"`
	KeyStore     string          `json:"key_store"`
	Chain        string          `json:"chain"`
	Gen          int             `json:"gen"`
	TokenId      string          `json:"token_id"`
	Price        decimal.Decimal `json:"price"`
	FormulaIndex int             `json:"formula_index"`
	Grade        int             `json:"grade"`
}

type TreasureSellInfo struct {
	Price       decimal.Decimal `json:"price"`
	Probability []int           `json:"probability"`
	PrizeRing   int             `json:"prize_ring"`
}

func (tq *TreasureQuery) TreasureList(ctx context.Context) ([]TreasureJson, map[string]int, []string) {
	var treasure []TreasureJson
	count := map[string]int{"goldCount": 0, "silverCount": 0, "lockCount": 0, "unlockCount": 0}

	query := util.WithContextDb(ctx).Model(Treasure{}).Where(tq.WhereQuery).Order("buy_time desc").Scan(&treasure)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil, count, nil
	}
	var tokenIds []string

	for i, v := range treasure {

		if v.Status == "locked" {
			treasure[i].KeyStore = ""
			count["lockCount"] = count["lockCount"] + 1
			if v.BoxType == "gold" {
				count["goldCount"] = count["goldCount"] + 1
			} else {
				count["silverCount"] = count["silverCount"] + 1
			}
			continue
		}
		tokenIds = append(tokenIds, v.TokenId)
		count["unlockCount"] = count["unlockCount"] + 1
	}
	maxLength := (tq.Page + 1) * tq.Row
	if maxLength > len(treasure) {
		maxLength = len(treasure)
	}
	if tq.Page*tq.Row >= len(treasure) {
		return nil, count, nil
	}
	treasure = treasure[tq.Page*tq.Row : maxLength]
	return treasure, count, tokenIds
}

func NewTreasure(txn *util.GormDB, buyer, tx, boxType string, price decimal.Decimal, blockTime, amount int64, chain string) error {
	var insertRecords []interface{}
	for i := 0; i < int(amount); i++ {
		insertRecords = append(insertRecords, Treasure{Buyer: buyer, Status: "locked",
			BoxIndex: genBox(util.U256(util.BytesToHex(uuid.Must(uuid.NewV4(), nil).Bytes())), boxType),
			BoxType:  boxType, TxId: tx, Price: price, Gen: 2, BuyTime: int(blockTime), IsLocked: true, Chain: chain})
	}
	return gormbulk.BulkInsert(txn.DB, insertRecords, 3000)
}

func (t *Treasure) OpenTreasure(ctx context.Context) error {
	db := util.DbBegin(ctx)
	defer db.DbRollback()

	if err := t.unlockTreasure(ctx, db); err != nil {
		return err
	}
	db.DbCommit()
	return nil
}

func (t *Treasure) unlockTreasure(ctx context.Context, db *util.GormDB) error {
	var minValue int
	if t.BoxType == "gold" {
		minValue = goldBoxMin + rand.Intn(goldBoxRandom)
	} else {
		minValue = silverBoxMin + rand.Intn(silverBoxRandom)
	}

	randomRing := decimal.RequireFromString(util.IntToString(minValue))
	randomCoo := randomRing.Mul(decimal.RequireFromString("10"))
	member := GetMember(ctx, int(t.MemberId))
	_ = member.TouchAccount(ctx, currencyRing, member.GetUseAddress(t.Chain), EthChain).AddBalance(db, randomRing, reasonUnlockTreasure)
	_ = member.TouchAccount(ctx, currencyCoo, member.GetUseAddress(t.Chain), EthChain).AddBalance(db, randomCoo, reasonUnlockTreasure)
	query := db.Model(&t).Where("id = ?", t.ID).UpdateColumn(Treasure{Status: "unlock", RingValue: randomRing, CooValue: randomCoo})

	if query.Error != nil {
		return errors.New("unlock treasure fail")
	}
	return nil
}

func (t *Treasure) ToJson() *TreasureJson {
	return &TreasureJson{
		BoxType:   t.BoxType,
		RingValue: t.RingValue,
		CooValue:  t.CooValue,
		BoxIndex:  t.BoxIndex,
		Status:    t.Status,
		KeyStore:  t.KeyStore,
	}
}

func (ec *EthTransactionCallback) LuckyBagCallback(ctx context.Context) error {
	if GetLt(ctx, ec.Tx) != nil {
		return errors.New("tx exist")
	}
	goldBoxCount := 0
	silverBoxCount := 0
	wallet := ""
	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("luckyBag", chain)) {
			logSlice := util.LogAnalysis(log.Data)
			if util.AddHex(log.Topics[0]) == services.AbiEncodingMethod("GoldBoxSale(address,uint256,uint256)") {
				wallet = util.AddHex(util.TrimHex(log.Topics[1])[24:64], chain)
				goldBoxCount = util.StringToInt(util.U256(logSlice[0]).String())
			} else if util.AddHex(log.Topics[0]) == services.AbiEncodingMethod("SilverBoxSale(address,uint256,uint256)") {
				wallet = util.AddHex(util.TrimHex(log.Topics[1])[24:64], chain)
				silverBoxCount = util.StringToInt(util.U256(logSlice[0]).String())
			}
		}
	}
	return updateTreasure(ctx, ec.Tx, wallet, goldBoxCount, silverBoxCount, chain)
}

func updateTreasure(ctx context.Context, tx, wallet string, goldBoxCount, silverBoxCount int, chain string) error {
	member := GetMemberByAddress(ctx, wallet, chain)
	memberId := uint(0)
	if member != nil {
		memberId = member.ID
	}
	db := util.DbBegin(ctx)
	defer db.DbRollback()
	var t Treasure
	if goldBoxCount > 0 {
		query := db.Model(&t).Where("box_type = ? and tx_id = ?", "gold", "").Limit(goldBoxCount).UpdateColumn(Treasure{MemberId: memberId, TxId: tx, BuyTime: int(time.Now().Unix()), Chain: chain})
		if query.Error != nil {
			return errors.New("unlock treasure fail")
		}
	}
	if silverBoxCount > 0 {
		query := db.Model(&t).Where("box_type = ? and tx_id = ?", "silver", "").Limit(silverBoxCount).UpdateColumn(Treasure{MemberId: memberId, TxId: tx, BuyTime: int(time.Now().Unix()), Chain: chain})
		if query.Error != nil {
			return errors.New("unlock treasure fail")
		}
	}
	lt := LuckyboxTrans{Tx: tx}
	if err := lt.New(ctx); err != nil {
		db.DbRollback()
		return err
	}
	db.DbCommit()
	return nil
}

//func GetTreasureByIndex(boxIndex string) *Treasure {
//	db := util.DB
//	var t Treasure
//	query := db.Where("box_index = ? ", boxIndex).First(&t)
//	if query == nil {
//		return nil
//	}
//	if query.Error != nil || !query.RecordNotFound() {
//		return nil
//	}
//	return &t
//}

func openGen2Treasure(txn *util.GormDB, boxIndex string, tokenId string, value decimal.Decimal, blockTime int64) error {
	query := txn.Model(Treasure{}).Where("box_index = ?", boxIndex).Update(map[string]interface{}{
		"status":     "unlock",
		"ring_value": value,
		"token_id":   tokenId,
		"open_time":  int(blockTime),
		"is_locked":  false,
	})
	return query.Error
}

func GetTreasureReversalKeyByIndex(ctx context.Context, ids []string, buyer string) map[string]Treasure {
	if len(ids) < 1 {
		return nil
	}
	db := util.WithContextDb(ctx)
	var treasures []Treasure
	query := db.Where("box_index in (?)", ids).Where("buyer = ?", buyer).Where("gen = 2 and status= 'locked'").Find(&treasures)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	results := make(map[string]Treasure)
	for _, v := range treasures {
		results[v.BoxIndex] = v
	}
	return results
}

func genBox(index *big.Int, boxType string) string {
	if boxType == "gold" {
		return util.BytesToHex(solsha3.U256(new(big.Int).Or(index, new(big.Int).Lsh(big.NewInt(1), 255))))
	}
	return util.BytesToHex(solsha3.U256(new(big.Int).Or(index, new(big.Int).Lsh(big.NewInt(1), 244))))
}
