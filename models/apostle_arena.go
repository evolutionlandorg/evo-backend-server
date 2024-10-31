package models

//
// import (
// 	"encoding/json"
// 	"github.com/evolutionlandorg/evo-backend/services"
// 	"github.com/evolutionlandorg/evo-backend/util"
// 	"errors"
// 	"fmt"
// 	"github.com/jinzhu/gorm"
// 	"github.com/shopspring/decimal"
// 	"time"
// )
//
// type ApostleArena struct {
// 	ID              uint `gorm:"primary_key"`
// 	CreatedAt       time.Time
// 	Defender        string `json:"defender"`
// 	DefenderAddress string `json:"defender_address"`
// 	DefenderName    string `json:"defender_name"`
// 	Attacker        string `json:"attacker"`
// 	AttackerAddress string `json:"attacker_address"`
// 	AttackerName    string `json:"attacker_name"`
// 	ArenaId         uint   `json:"arena_id"`
// 	Winner          string `json:"winner"`
// 	WinnerApostle   string `json:"winner_apostle"`
// 	WinnerName      string `json:"winner_name"`
// 	CompetitionItem string `json:"competition_item"`
// 	CompetitionSide string `json:"competition_side"`
// 	StartAt         int    `json:"start_at"`
// 	CreateTx        string `json:"create_tx"`
// 	FinishTx        string `json:"finish_tx"`
// 	Status          string `json:"status"` // going,stop,finish
// 	Chain           string `json:"chain"`
// }
//
// type ApostleArenaRecord struct {
// 	ID             uint `gorm:"primary_key"`
// 	CreatedAt      time.Time
// 	ArenaId        uint            `json:"arena_id"`
// 	ApostleArenaId uint            `json:"apostle_arena_id"`
// 	ParticipateTx  string          `json:"participate_tx"`
// 	Address        string          `json:"address"`
// 	Name           string          `json:"name"`
// 	Result         string          `json:"result"`
// 	Choose         string          `json:"choose"`
// 	BetAmount      decimal.Decimal `json:"bet_amount" sql:"type:decimal(32,18);"`
// 	WinAmount      decimal.Decimal `json:"win_amount" sql:"type:decimal(32,18);"`
// }
//
// type ApostleArenaHistory struct {
// 	ID             uint            `gorm:"primary_key" json:"id"`
// 	CreatedAt      time.Time       `json:"created_at"`
// 	ArenaId        uint            `json:"arena_id"`
// 	ApostleArenaId uint            `json:"apostle_arena_id"`
// 	Address        string          `json:"address"`
// 	Type           string          `json:"type"`
// 	AddTime        int             `json:"add_time"`
// 	Amount         decimal.Decimal `json:"amount" sql:"type:decimal(32,18);"`
// }
//
// type ApostleArenaRank struct {
// 	Address string          `json:"address"`
// 	Name    string          `json:"name"`
// 	Bonus   decimal.Decimal `json:"bonus" sql:"type:decimal(32,18);"`
// }
//
// type ApostleGameJson struct {
// 	Owner          string `json:"owner"`
// 	Introduction   string `json:"introduction" sql:"type:text;"`
// 	TokenId        string `json:"token_id"`
// 	TokenIndex     int    `json:"token_index"`
// 	ApostlePicture string `json:"apostle_picture"`
// 	Gen            int    `json:"gen"`
// 	Name           string `json:"name"`
// 	Gender         string `json:"gender"`
// 	*ApostleTalentJson
// 	District int `json:"district"`
// }
//
// type ApostleArenaJson struct {
// 	Defender        *ApostleGameJson   `json:"defender"`
// 	Attacker        *ApostleGameJson   `json:"attacker"`
// 	ArenaId         uint               `json:"arena_id"`
// 	Winner          *ApostleGameJson   `json:"winner"`
// 	CompetitionItem string             `json:"competition_item"`
// 	CompetitionSide string             `json:"competition_side"`
// 	StartAt         int                `json:"start_at"`
// 	Status          string             `json:"status"` // going,stop,finish
// 	DefenderBetList []ArenaBetList     `json:"defender_bet_list"`
// 	AttackerBetList []ArenaBetList     `json:"attacker_bet_list"`
// 	BonusList       []ApostleArenaRank `json:"bonus_list"`
// }
// type ArenaBetList struct {
// 	Address   string          `json:"address"`
// 	Name      string          `json:"name"`
// 	Choose    string          `json:"choose"`
// 	BetAmount decimal.Decimal `json:"bet_amount" sql:"type:decimal(32,16);"`
// 	WinAmount decimal.Decimal `json:"win_amount" sql:"type:decimal(32,16);"`
// }
//
// var (
// 	CompetitionItemMap = map[string]string{"0": "strength", "1": "agile", "2": "intellect", "3": "hp", "4": "mood", "5": "finesse", "6": "lucky", "7": "charm"}
// 	CompetitionSideMap = map[string]string{"0": "less", "1": "great"}
// 	ArenaRoundTime     = map[string]int{"dev": 60 * 20, "production": 60 * 120}
// )
//
// func (aa *ApostleArena) New(txn *util.GormDB) error {
// 	aa.Status = "going"
// 	result := txn.Create(&aa)
// 	return result.Error
// }
//
// func (ar *ApostleArenaRecord) New(txn *util.GormDB) error {
// 	result := txn.Create(&ar)
// 	return result.Error
// }
//
// func (apq *ApostleQuery) ApostlesByAddress() ([]ApostleGameJson, int) {
// 	db := util.DB
// 	var (
// 		apostles []ApostleGameJson
// 		count    int
// 	)
// 	query := db.Table("apostles").Select("apostles.*,apostle_talents.*").Where(apq.WhereQuery).
// 		Joins("join apostle_talents on apostles.id=apostle_talents.apostle_id").
// 		Offset(apq.Page * apq.Row).Limit(apq.Row).
// 		Order("apostles.id asc").Scan(&apostles)
// 	if apq.Row < len(apostles) {
// 		count = len(apostles)
// 	} else {
// 		db.Table("apostles").Where(apq.WhereQuery).Count(&count)
// 	}
// 	if query.Error != nil || query == nil || query.RecordNotFound() {
// 		return nil, 0
// 	}
// 	return apostles, count
// }
//
// func (ec *EthTransactionCallback) ApostleArenaCallback() (err error) {
// 	if getTransactionDeal(ec.Tx, "apostleArena") != nil {
// 		return errors.New("tx exist")
// 	}
// 	txn := util.DbBegin(context.TODO())
// 	defer txn.DbRollback()
// 	chain := ec.Receipt.ChainSource
// 	execute := false
// 	for _, log := range ec.Receipt.Logs {
// 		if len(log.Topics) != 0 && util.AddHex(log.Address, chain) == util.GetContractAddress("apostleArena", chain) {
// 			eventName := util.AddHex(log.Topics[0])
// 			dataSlice := services.LogAnalysis(log.Data)
// 			switch eventName {
// 			case services.AbiEncodingMethod("RegRound(uint256,uint256,uint256,uint256)"): // 开局
// 			case services.AbiEncodingMethod("PlayerJoin(address,uint256,uint256,uint256,uint256)"): // 参与
// 				execute = true
// 				arenaId := uint(util.StringToInt(util.U256(dataSlice[2]).String()))
// 				address := util.AddHex(dataSlice[0][24:64], chain)
// 				tokenId := dataSlice[1]
// 				startAt := util.StringToInt(util.U256(dataSlice[3]).String())
// 				aa := GetArenaByArenaId(uint(arenaId), chain)
// 				saveFlag := false
// 				if aa == nil {
// 					aa = &ApostleArena{Chain: chain, CreateTx: ec.Tx, StartAt: startAt, ArenaId: arenaId}
// 					saveFlag = true
// 				}
// 				if util.StringToInt(util.U256(dataSlice[4]).String()) == 0 {
// 					aa.Defender = tokenId
// 					aa.DefenderAddress = address
// 					if defenderMember := GetMemberByAddress(aa.DefenderAddress, chain); defenderMember != nil {
// 						aa.DefenderName = defenderMember.Name
// 					}
// 				} else {
// 					aa.Attacker = tokenId
// 					aa.AttackerAddress = address
// 					if attackerMember := GetMemberByAddress(aa.AttackerAddress, chain); attackerMember != nil {
// 						aa.AttackerName = attackerMember.Name
// 					}
// 				}
// 				if saveFlag {
// 					err = aa.New(txn)
// 				} else {
// 					err = txn.Save(&aa).Error
// 				}
// 				if err == nil {
// 					aa.addHistory(txn, address, "join", decimal.New(5000, 0).Neg())
// 				}
// 			case services.AbiEncodingMethod("SettleRound(uint256,address,uint256,uint256,uint256,uint8,bool)"): // 结算
// 				execute = true
// 				arenaId := uint(util.StringToInt(util.U256(dataSlice[0]).String()))
// 				if aa := GetArenaByArenaId(uint(arenaId), chain); aa != nil {
// 					aa.Winner = util.AddHex(dataSlice[1][24:64], chain)
// 					aa.WinnerApostle = dataSlice[4]
// 					if aa.Winner == aa.DefenderAddress {
// 						aa.WinnerName = aa.DefenderName
// 					} else if aa.Winner == aa.AttackerAddress {
// 						aa.WinnerName = aa.AttackerName
// 					}
// 					aa.CompetitionItem = CompetitionItemMap[util.U256(dataSlice[5]).String()]
// 					aa.CompetitionSide = CompetitionSideMap[util.U256(dataSlice[6]).String()]
// 					aa.FinishTx = ec.Tx
// 					aa.Status = "finish"
// 					if err = txn.Save(&aa).Error; err == nil {
// 						err = aa.distribution(util.BigToDecimal(util.U256(dataSlice[2])), util.BigToDecimal(util.U256(dataSlice[3]))) // 分发奖励
// 					}
// 					util.DelCache("apostle:arena:Rank")
// 				} else {
// 					err = errors.New(fmt.Sprintf("get find this arena %d", arenaId))
// 				}
// 			case services.AbiEncodingMethod("InvalidRound(uint256)"): // 流局
// 				execute = true
// 				arenaId := uint(util.StringToInt(util.U256(dataSlice[0]).String()))
// 				if aa := GetArenaByArenaId(uint(arenaId), chain); aa != nil && aa.Status == "going" {
// 					aa.Status = "stop"
// 					aa.FinishTx = ec.Tx
// 					if err = txn.Save(&aa).Error; err == nil {
// 						wallet := aa.DefenderAddress
// 						if wallet == "" {
// 							wallet = aa.AttackerAddress
// 						}
// 						err = addRewardOnce(txn, wallet, ReasonApostleArenaSendBack, ec.Tx, decimal.New(5000, 0), chain, currencyRing)
// 						aa.addHistory(txn, wallet, "joinSendBack", decimal.New(5000, 0))
// 					}
// 				} else {
// 					err = errors.New(fmt.Sprintf("get find this arena %d", arenaId))
// 				}
// 				execute = true
// 			case services.AbiEncodingMethod("FollowerJoin(address,address,uint256,uint256,uint256)"): // 投注
// 				execute = true
// 				ar := ApostleArenaRecord{Address: util.AddHex(dataSlice[0][24:64], chain), ParticipateTx: ec.Tx, ArenaId: uint(util.StringToInt(util.U256(dataSlice[4]).String()))}
// 				if aa := GetArenaByArenaId(ar.ArenaId, chain); aa != nil {
// 					ar.ApostleArenaId = aa.ID
// 					ar.Choose = dataSlice[2]
// 					ar.BetAmount = util.BigToDecimal(util.U256(dataSlice[3]))
// 					if member := GetMemberByAddress(ar.Address, chain); member != nil {
// 						ar.Name = member.Name
// 					}
// 					if err = ar.New(txn); err == nil {
// 						aa.addHistory(txn, ar.Address, "bet", ar.BetAmount.Neg())
// 					}
// 				} else {
// 					err = errors.New(fmt.Sprintf("get find this arena %d", ar.ArenaId))
// 				}
// 			}
// 		}
// 	}
// 	if !execute || err != nil {
// 		return errors.New(fmt.Sprintf("get execute %v or error %s", execute, err))
// 	}
// 	if err := ec.NewUniqueTransaction("apostleArena"); err != nil {
// 		return err
// 	}
// 	txn.DbCommit()
// 	return err
// }
//
// func (aa *ApostleArena) GetJoinBetList(choose string) []ApostleArenaRecord {
// 	db := util.DB
// 	var list []ApostleArenaRecord
// 	if aa.Winner == util.GetContractAddress("apostleArena", aa.Chain) {
// 		db.Table("apostle_arena_records").Where("apostle_arena_id = ?", aa.ID).Find(&list)
// 	} else {
// 		db.Table("apostle_arena_records").Where("apostle_arena_id = ?", aa.ID).Where("choose = ?", choose).Find(&list)
// 	}
// 	return list
// }
//
// func (aa *ApostleArena) distribution(joinAmount, betAmount decimal.Decimal) error {
// 	if joinAmount.Sign() > 0 {
// 		txn := util.DbBegin(context.TODO())
// 		if err := addRewardOnce(txn, aa.Winner, ReasonApostleArenaWin, aa.FinishTx, joinAmount, aa.Chain, currencyRing); err != nil {
// 			return err
// 		}
// 		aa.addHistory(txn, aa.Winner, "win", joinAmount)
// 		txn.DbCommit()
// 	}
// 	if betAmount.Sign() <= 0 {
// 		return nil
// 	}
// 	total := aa.getTotalBetAmount()
// 	if total.Sign() <= 0 {
// 		return nil
// 	}
// 	for _, v := range aa.GetJoinBetList(aa.WinnerApostle) {
// 		reward := v.BetAmount.Div(total).Mul(betAmount)
// 		txn := util.DbBegin(context.TODO())
// 		addRewardOnce(txn, v.Address, ReasonApostleArenaBetWin, aa.FinishTx, reward, aa.Chain, currencyRing)
// 		txn.Model(&v).UpdateColumn(ApostleArenaRecord{WinAmount: reward, Result: "win"})
// 		aa.addHistory(txn, v.Address, "winBet", reward)
// 		txn.DbCommit()
// 	}
// 	return nil
// }
//
// func (aa *ApostleArena) addHistory(db *util.GormDB, address, action string, amount decimal.Decimal) error {
// 	ah := ApostleArenaHistory{ArenaId: aa.ArenaId, ApostleArenaId: aa.ID, Address: address, AddTime: int(time.Now().Unix()), Type: action, Amount: amount}
// 	result := db.Create(&ah)
// 	return result.Error
// }
//
// func (aa *ApostleArena) getTotalBetAmount() decimal.Decimal {
// 	db := util.DB
// 	var total StatSum
// 	var query *gorm.DB
// 	if aa.Winner == util.GetContractAddress("apostleArena", aa.Chain) {
// 		query = db.Table("apostle_arena_records").Select("sum(bet_amount) as total").Where("apostle_arena_id = ?", aa.ID).Scan(&total)
// 	} else {
// 		query = db.Table("apostle_arena_records").Select("sum(bet_amount) as total").Where("apostle_arena_id = ? and choose = ?", aa.ID, aa.WinnerApostle).Scan(&total)
// 	}
// 	if query == nil || query.Error != nil {
// 		return decimal.Zero
// 	}
// 	return total.Total
// }
//
// func (aa *ApostleArena) AsJson() *ApostleArenaJson {
// 	if aa == nil {
// 		return nil
// 	}
// 	var aj ApostleArenaJson
// 	aj.ArenaId = aa.ArenaId
// 	aj.CompetitionSide = aa.CompetitionSide
// 	aj.CompetitionItem = aa.CompetitionItem
// 	aj.StartAt = aa.StartAt
// 	aj.Status = aa.Status
// 	aj.Defender = apostlesGameDisplay(aa.Defender)
// 	aj.Attacker = apostlesGameDisplay(aa.Attacker)
//
// 	if aj.Attacker != nil {
// 		db := util.DB
// 		var list []ArenaBetList
// 		db.Table("apostle_arena_records").Where("apostle_arena_id = ?", aa.ID).Scan(&list)
// 		for _, v := range list {
// 			if v.Choose == aa.Defender {
// 				aj.DefenderBetList = append(aj.DefenderBetList, v)
// 			} else {
// 				aj.AttackerBetList = append(aj.AttackerBetList, v)
// 			}
// 		}
// 	}
// 	aj.Attacker = apostlesGameDisplay(aa.Attacker)
// 	aj.Winner = nil
// 	if aj.Status == "finish" {
// 		if aa.Winner == aa.DefenderAddress {
// 			aj.Winner = aj.Defender
// 		} else if aa.Winner == aa.AttackerAddress {
// 			aj.Winner = aj.Attacker
// 		}
// 		aj.BonusList = aa.getBonusList()
// 	}
// 	return &aj
// }
//
// func (aa *ApostleArena) getBonusList() []ApostleArenaRank {
// 	db := util.DB
// 	var rank []ApostleArenaRank
// 	cacheKey := fmt.Sprintf("apostle:arena:round:%d", aa.ID)
// 	if rankBytes := util.GetCache(cacheKey); rankBytes == nil {
// 		query := db.Table("apostle_arena_histories").Select("address, sum(amount) as bonus").Where("apostle_arena_id = ?", aa.ID).
// 			Where("type in(?)", []string{"winBet", "win"}).Group("address").Order("bonus desc").Scan(&rank)
// 		if query == nil || query.Error != nil || query.RecordNotFound() || len(rank) == 0 {
// 			return nil
// 		}
// 		for k, v := range rank {
// 			if m := GetMemberByAddress(v.Address, aa.Chain); m != nil {
// 				rank[k].Name = m.Name
// 			}
// 		}
// 		if rankBytes, err := json.Marshal(rank); err == nil {
// 			util.SetCache(cacheKey, rankBytes, 10*60)
// 		}
// 	} else {
// 		json.Unmarshal(rankBytes, &rank)
// 	}
// 	return rank
// }
//
// func GetCurrentArena(chain string) *ApostleArena {
// 	db := util.DB
// 	var arena ApostleArena
// 	t := int64(ArenaRoundTime[util.Environment])
// 	startAt := (time.Now().Unix() / t) * t
// 	query := db.Model(ApostleArena{}).Where("start_at = ? and chain = ?", startAt, chain).First(&arena)
// 	if query.Error != nil || query.RecordNotFound() {
// 		return nil
// 	}
// 	return &arena
// }
//
// func GetArenaByArenaId(arenaId uint, chain string) *ApostleArena {
// 	db := util.DB
// 	var arena ApostleArena
// 	query := db.Model(ApostleArena{}).Where("arena_id = ? and chain = ?", arenaId, chain).First(&arena)
// 	if query.Error != nil || query.RecordNotFound() {
// 		return nil
// 	}
// 	return &arena
// }
//
// func GetLastArena(chain string) map[string]interface{} {
// 	t := int64(ArenaRoundTime[util.Environment])
// 	startAt := (time.Now().Unix() / t) * t
// 	if time.Now().Unix()-startAt > 10*60 {
// 		return nil
// 	}
// 	db := util.DB
// 	var arena ApostleArena
// 	query := db.Model(ApostleArena{}).Where("start_at = ? and chain = ?", startAt-t, chain).First(&arena)
// 	if query.Error != nil || query.RecordNotFound() {
// 		return nil
// 	}
// 	seat := 0
// 	if arena.Winner == arena.AttackerAddress {
// 		seat = 1
// 	}
// 	return map[string]interface{}{"winner": arena.Winner, "winner_name": arena.WinnerName, "seat": seat}
// }
//
// func apostlesGameDisplay(tokenId string) *ApostleGameJson {
// 	if tokenId == "" {
// 		return nil
// 	}
// 	db := util.DB
// 	var apostle ApostleGameJson
// 	query := db.Table("apostles").Select("apostles.*,apostle_talents.*").Where("apostles.token_id = ?", tokenId).
// 		Joins("join apostle_talents on apostles.id=apostle_talents.apostle_id").Scan(&apostle)
// 	if query == nil || query.Error != nil || query.RecordNotFound() {
// 		return nil
// 	}
// 	return &apostle
// }
//
// func GetArenaHistory(address string, page, row int) ([]ApostleArenaHistory, int) {
// 	var history []ApostleArenaHistory
// 	db := util.DB
// 	var count int
// 	db.Model(ApostleArenaHistory{}).Where("address = ?", address).Count(&count)
// 	if count == 0 {
// 		return []ApostleArenaHistory{}, 0
// 	}
// 	query := db.Model(ApostleArenaHistory{}).Where("address = ?", address).Offset(page * row).Limit(row).
// 		Order("id desc").Find(&history)
// 	if query == nil || query.Error != nil || query.RecordNotFound() {
// 		return []ApostleArenaHistory{}, 0
// 	}
// 	return history, count
// }
//
// func GetArenaRank(chain string) []ApostleArenaRank {
// 	var rank []ApostleArenaRank
// 	if rankBytes := util.GetCache("apostle:arena:Rank"); rankBytes == nil {
// 		db := util.DB
// 		query := db.Table("apostle_arena_histories").Select("address,type, sum(amount) as bonus").Where("type in(?)", []string{"winBet", "win"}).Group("address").
// 			Order("bonus desc").Scan(&rank)
// 		if query == nil || query.Error != nil || query.RecordNotFound() || len(rank) == 0 {
// 			return nil
// 		}
// 		for k, v := range rank {
// 			if m := GetMemberByAddress(v.Address, chain); m != nil {
// 				rank[k].Name = m.Name
// 			}
// 		}
// 		if rankBytes, err := json.Marshal(rank); err == nil {
// 			util.SetCache("apostle:arena:Rank", rankBytes, 60*115)
// 		}
//
// 	} else {
// 		json.Unmarshal(rankBytes, &rank)
// 	}
// 	return rank
// }
