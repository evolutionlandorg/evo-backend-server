package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type Snapshot struct {
	gorm.Model    `json:"-"`
	Chain         string          `json:"chain"`
	Wallet        string          `json:"wallet"`
	LandNumber    int64           `json:"land_number"`
	ApostleNumber int64           `json:"apostle_number"`
	KtonNumber    decimal.Decimal `json:"kton_number" sql:"type:decimal(36,18);"`
	Timestamp     int64           `json:"timestamp"`
}

type SnapshotReq struct {
	Options   SnapshotReqOptions `json:"options"`
	Network   string             `json:"network"`
	Addresses []string           `json:"addresses"`
	Snapshot  interface{}        `json:"snapshot"`
}

type SnapshotReqOptions struct {
	Land    int64    `json:"land"`    // land vote score
	Apostle int64    `json:"apostle"` // apostle vote score
	Kton    float64  `json:"kton"`    // kton vote score
	Element float64  `json:"element"` // element vote score
	Chain   []string `json:"chain"`
}

type SnapshotAsJson struct {
	Score []Score `json:"score"`
}

type Score struct {
	Score   int64  `json:"score"`
	Address string `json:"address"`
}

type SnapshotResp struct {
	Land    int64           `json:"land"`
	Apostle int64           `json:"apostle"`
	Kton    decimal.Decimal `json:"kton"`
}

func (s Snapshot) TableName() string {
	return "snapshot"
}

func GetSnapshotByAddress(ctx context.Context, address []string, blockTag int, chain []string, useBlockChain string) (map[string]*Snapshot, error) {
	sg := storage.New(useBlockChain)
	var blockTimestamp int64
	if blockTag != 0 {
		if block := sg.BlockHeader(uint64(blockTag)); block != nil {
			blockTimestamp = int64(block.BlockTimeStamp)
		}
	}
	now := time.Now().Unix()
	if blockTimestamp > now {
		blockTimestamp = now
	}
	db := util.WithContextDb(ctx)
	if blockTimestamp != 0 {
		db = db.Raw(`SELECT * from snapshot WHERE id IN ((SELECT MAX(id) as id FROM snapshot
					WHERE ((timestamp <= ?) AND (wallet IN (?) AND chain IN (?))) 
					GROUP BY wallet, chain ORDER BY MAX(id) DESC))`, blockTimestamp, address, chain)
	} else {
		db = db.Raw(`SELECT * from snapshot WHERE id IN ((SELECT MAX(id) as id FROM snapshot
					WHERE (wallet IN (?) AND chain IN (?))
					GROUP BY wallet, chain ORDER BY MAX(id) DESC))`, address, chain)
	}

	var snapshots []*Snapshot

	if err := db.Scan(&snapshots).Error; err != nil {
		return nil, err
	}
	var result = make(map[string]*Snapshot, len(snapshots))
	for index, v := range snapshots {
		if _, ok := result[strings.ToLower(v.Wallet)]; !ok {
			result[strings.ToLower(v.Wallet)] = snapshots[index]
			continue
		}
		result[strings.ToLower(v.Wallet)].LandNumber += snapshots[index].LandNumber
		result[strings.ToLower(v.Wallet)].ApostleNumber += snapshots[index].ApostleNumber
	}

	return result, nil
}

func GetLatestSnapshot(ctx context.Context, chain string) (map[string]*Snapshot, error) {
	db := util.WithContextDb(ctx)
	db = db.Raw(`SELECT * from snapshot WHERE id IN ((SELECT MAX(id) as id FROM snapshot WHERE chain = ? GROUP BY wallet, chain ORDER BY MAX(id) DESC))`, chain)
	var snapshots []*Snapshot

	if err := db.Scan(&snapshots).Error; err != nil {
		return nil, err
	}
	var result = make(map[string]*Snapshot, len(snapshots))
	for index, v := range snapshots {
		owner := strings.ToLower(v.Wallet)
		if _, ok := result[owner]; !ok {
			result[owner] = snapshots[index]
			continue
		}
		result[owner].LandNumber += snapshots[index].LandNumber
		result[owner].ApostleNumber += snapshots[index].ApostleNumber
	}
	return result, nil
}

type OwnerAssert struct {
	Owner string `json:"owner"`
	Count int64  `json:"count"`
}

type OwnerVote struct {
	LandNumber    int64 `json:"land_number"`
	ApostleNumber int64 `json:"apostle_number"`
	KtonNumber    int64 `json:"kton_number"`
}

func SaveSnapshot(ctx context.Context, chain string) func() {
	return func() {
		if util.IncrCache(ctx, fmt.Sprintf("SaveSnapshot%s", chain), 60) > 1 {
			return
		}
		var chains = []string{chain}
		if util.StringInSlice(EthChain, chains) {
			chains = append(chains, "ethereum")
		}
		now := time.Now().Unix()
		var ownerApostle []OwnerAssert
		db := util.WithContextDb(ctx).Table("apostles").Where("chain in (?)", chains).Group("owner,chain")
		if err := db.Select(`count(*) as count, owner`).Find(&ownerApostle).Error; err != nil {
			log.Error("find apostle error %s", err)
			return
		}
		// get owner land number
		var ownerLand []OwnerAssert
		db = util.WithContextDb(ctx).Table("lands").Where("chain in (?)", chains).Group("owner,chain")
		if err := db.Select(`count(*) as count, owner`).Find(&ownerLand).Error; err != nil {
			log.Error("find lands error %s", err)
			return
		}

		var owners = make(map[string]*OwnerVote)
		for _, v := range ownerApostle {
			owner := strings.ToLower(v.Owner)
			if _, ok := owners[owner]; !ok {
				owners[owner] = &OwnerVote{}
			}

			owners[owner].ApostleNumber += v.Count
		}

		for _, v := range ownerLand {
			owner := strings.ToLower(v.Owner)
			if _, ok := owners[owner]; !ok {
				owners[owner] = &OwnerVote{}
			}
			owners[owner].LandNumber += v.Count
		}

		latest, _ := GetLatestSnapshot(ctx, chain)
		// save
		txn := util.DbBegin(ctx)
		defer txn.DbRollback()
		for wallet, v := range owners {
			if _, ok := latest[wallet]; ok && v.LandNumber == latest[wallet].LandNumber && v.ApostleNumber == latest[wallet].ApostleNumber {
				continue
			}
			if err := txn.Create(&Snapshot{
				Chain:         chain,
				Wallet:        wallet,
				LandNumber:    v.LandNumber,
				ApostleNumber: v.ApostleNumber,
				Timestamp:     now,
			}).Error; err != nil {
				log.Error("create snapshot error %s", err)
				return
			}
		}
		txn.DbCommit()
		if err := txn.Error; err != nil {
			log.Error("create snapshot error %s", err)
			return
		}
	}
}
