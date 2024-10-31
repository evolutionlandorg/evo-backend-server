// Package models provides ...
package models

import (
	"context"
	"math/rand"
	"time"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/crypto"

	"github.com/jinzhu/gorm"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"
)

type PloLeft struct {
	gorm.Model `json:"-"`
	Origin     string `json:"origin"`
	Prize      int    `json:"prize"`
	Total      int    `json:"total"`
	Stock      int    `json:"stock"`
}

type PRate struct {
	Prize int `json:"prize"`
	CodeA int `json:"code_a"`
	CodeB int `json:"code_b"`
}

type PloTicket struct {
	gorm.Model `json:"-"`
	Origin     string `json:"origin"`
	PubKey     string `json:"pub_key"`
	N          int    `json:"n"`
}

type PloRaffleRecord struct {
	gorm.Model `json:"-"`
	Origin     string         `json:"origin"`
	PubKey     string         `json:"pub_key"`
	Sign       string         `json:"sign"`
	N          int            `json:"n"`
	Prize      int            `json:"prize"`
	Land       int            `json:"land"`
	Apostle    int            `json:"apostle"`
	Box        string         `json:"box"`
	To         string         `json:"to"`
	Btc        string         `json:"btc"`
	KeyTypeID  crypto.KeyType `json:"key_type_id"`
}

type PloTotal struct {
	Origin string `json:"origin"`
	Total  int    `json:"total"`
}

type Reward struct {
	N       int    `json:"n"`
	Prize   int    `json:"prize"`
	Land    int    `json:"land"`
	Apostle int    `json:"apostle"`
	Box     string `json:"box"`
}

const (
	Empty = iota
	Rare
	Epic
	Legendary
)

const (
	EmptyBox  = "empty"
	SilverBox = "silver"
	GoldBox   = "gold"
)

func PloRaffle(ctx context.Context, origin string, now int, n int, pubKey, sign, to, btc string, keyTypeID crypto.KeyType) (*[]Reward, error) {
	txn := util.DbBegin(ctx)
	defer txn.DbRollback()
	ploRates := RawPloLeft(ctx, origin)
	rates := make(map[int]int)
	for _, r := range ploRates {
		rates[r.Prize] = r.Stock
	}
	rewards, newRates := doRaffle(now, n, rates)
	for _, r := range ploRates {
		oldLeft := r.Stock
		newLeft := newRates[r.Prize]
		if oldLeft != newLeft {
			if err := RawUpdatePloLeft(txn, origin, r.Prize, oldLeft, newLeft); err != nil {
				return nil, err
			}
		}
	}
	if err := RawAddPloRaffleRecords(txn, origin, pubKey, sign, to, btc, keyTypeID, rewards); err != nil {
		return nil, err
	}
	txn.DbCommit()
	return rewards, nil
}

func doRaffle(now int, n int, rates map[int]int) (*[]Reward, map[int]int) {
	prizes := make([]Reward, n)
	for i := 0; i < n; i++ {
		myPrize := Empty
		nonce := now + i + 1
		total := 0
		for _, v := range rates {
			total += v
		}
		if total > 0 {
			code := random(total)
			prev := 0
			for k, v := range rates {
				if v > 0 {
					if code >= prev && code < prev+v {
						myPrize = k
						break
					} else {
						prev += v
					}
				}
			}
			prizes[i] = newPrize(nonce, myPrize)
		} else {
			prizes[i] = newPrize(nonce, myPrize)
		}
		rates[myPrize] = rates[myPrize] - 1
	}
	return &prizes, rates
}

func random(seed int) int {
	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	return r.Intn(seed)
}

func newPrize(n int, prize int) Reward {
	var reward Reward
	reward.N = n
	reward.Prize = prize
	code := random(100)
	if code%100 < 10 {
		reward.Box = GoldBox
	} else {
		reward.Box = SilverBox
	}
	switch prize {
	case Empty:
		reward.Land = 0
		reward.Apostle = 0
		reward.Box = EmptyBox
	case Rare:
		reward.Land = 0
		reward.Apostle = 0
	case Epic:
		reward.Land = 0
		reward.Apostle = 1
	case Legendary:
		reward.Land = 1
		reward.Apostle = 1
	}
	return reward
}

func RawPloLeft(ctx context.Context, origin string) []PloLeft {
	var rates []PloLeft
	util.WithContextDb(ctx).Where("origin = ?", origin).Order("prize asc").Find(&rates)
	return rates
}

func RawUpdatePloLeft(txn *util.GormDB, origin string, prize, oldLeft, newLeft int) error {
	return txn.Model(&PloLeft{}).Where("origin = ?", origin).Where("prize = ?", prize).Where("stock = ?", oldLeft).Update("stock", newLeft).Error
}

func RawAddPloRaffleRecords(txn *util.GormDB, origin, pubKey, sign, to, btc string, keyTypeId crypto.KeyType, prizes *[]Reward) error {
	var records []interface{}
	for _, prize := range *prizes {
		records = append(records, PloRaffleRecord{
			Origin:    origin,
			PubKey:    pubKey,
			Sign:      sign,
			N:         prize.N,
			Prize:     prize.Prize,
			Land:      prize.Land,
			Apostle:   prize.Apostle,
			Box:       prize.Box,
			To:        to,
			Btc:       btc,
			KeyTypeID: keyTypeId,
		})
	}
	return gormbulk.BulkInsert(txn.DB, records, 2000)
}

func RawPloRaffleRecordPrizes(ctx context.Context, origin string, prize int) int {
	var count int
	util.WithContextDb(ctx).Model(PloRaffleRecord{}).Where("origin = ?", origin).Where("prize = ?", prize).Count(&count)
	return count
}

func RawMaxPloRaffleRecord(ctx context.Context, origin string, pubKey string) PloRaffleRecord {
	var record PloRaffleRecord
	util.WithContextDb(ctx).Where("origin = ?", origin).Where("pub_key = ?", pubKey).Order("n desc").First(&record)
	return record
}

func RawPloRaffleLandRecords(ctx context.Context, page int, row int) []PloRaffleRecord {
	var records []PloRaffleRecord
	util.WithContextDb(ctx).Offset(page*row).Where("land = ?", 1).Order("id desc").Limit(row).Find(&records)
	return records
}

func RawPloRaffleRecords(ctx context.Context, page int, row int, origin string, pubKey string) []PloRaffleRecord {
	var records []PloRaffleRecord
	util.WithContextDb(ctx).Offset(page*row).Where("origin = ?", origin).Where("pub_key = ?", pubKey).Order("n desc").Limit(row).Find(&records)
	return records
}

func RawPloTicket(ctx context.Context, origin string, pubKey string) *PloTicket {
	var ticket PloTicket
	util.WithContextDb(ctx).Where("origin = ?", origin).Where("pub_key = ?", pubKey).First(&ticket)
	return &ticket
}
