package models

import (
	"context"
	"errors"

	"github.com/evolutionlandorg/evo-backend/util"
)

func MigrationDbTable() error {
	ctx := context.TODO()
	if util.WithContextDb(ctx) == nil {
		return errors.New("db not init")
	}
	db := util.WithContextDb(ctx)
	db.Set("gorm:table_options", "ENGINE=InnoDB").
		AutoMigrate(
			&Member{},
			&ElementRaffle{},
			&Account{},
			&Withdraw{},
			&Chat{},
			&Land{},
			&AccountVersion{},
			&TransactionScan{},
			&Auction{},
			&EthTransaction{},
			&LandData{},
			&Treasure{},
			&AuctionHistory{},
			&LuckyboxTrans{},
			&BroadcastMessage{},
			&TransactionHistory{},
			&UniqueTransaction{},
			&KeyStore{},
			&Apostle{},
			&LandApostle{},
			&Attribute{},
			&ApostleAttribute{},
			&ApostleTalent{},
			&AuctionApostle{},
			&ApostleWorkTrade{},
			&ApostleFertility{},
			&ApostlePregnant{},
			&ApostleReward{},
			&TokenSwap{},
			&ApostlePet{},
			&PetMirror{},
			&Dapp{},
			&DappReport{},
			// &ApostleArena{},
			// &ApostleArenaRecord{},
			// &ApostleArenaHistory{},
			&Building{},
			&BuildingAdminHiring{},
			&BuildingAdmin{},
			&BuildingHiring{},
			&BuildingWorker{},
			&MemberTakeBack{},
			&Drill{},
			&LandEquip{},
			&PloLeft{},
			&PloTicket{},
			&PloRaffleRecord{},
			&FarmAPR{},
			&Equipment{},
			&ParseTxError{},
			&MemberLoginInfo{},
		)

	db.Model(Account{}).AddIndex("member_currency", "member_id", "currency")
	db.Model(Withdraw{}).AddIndex("account_id", "account_id")
	db.Model(Withdraw{}).AddIndex("member_id", "member_id")
	db.Model(Withdraw{}).AddUniqueIndex("tx_id", "tx_id")
	db.Model(Chat{}).AddIndex("member_id", "member_id")
	db.Model(Land{}).AddIndex("member_id", "member_id")
	db.Model(Land{}).AddIndex("owner", "owner")
	db.Model(Land{}).AddUniqueIndex("lon_lat", "lon", "lat")
	db.Model(Land{}).AddUniqueIndex("token_id", "token_id")
	db.Model(Land{}).AddIndex("district", "district")
	db.Model(Auction{}).AddIndex("token_id", "token_id")
	db.Model(Auction{}).AddUniqueIndex("create_tx", "create_tx")
	db.Model(AuctionHistory{}).AddIndex("tx_id", "tx_id")
	db.Model(EthTransaction{}).AddUniqueIndex("tx", "tx")
	db.Model(LuckyboxTrans{}).AddUniqueIndex("tx", "tx")
	db.Model(LandData{}).AddUniqueIndex("token_id", "token_id")
	db.Model(LandData{}).AddUniqueIndex("land_id", "land_id")
	db.Model(Treasure{}).AddIndex("box_index", "box_index")
	db.Model(Treasure{}).AddIndex("buyer", "buyer")
	db.Model(Treasure{}).AddIndex("status", "status")
	db.Model(UniqueTransaction{}).AddUniqueIndex("tx", "tx", "action")
	db.Model(UniqueTransaction{}).AddIndex("confirm_chain", "confirm", "chain", "block_num")
	db.Model(KeyStore{}).AddUniqueIndex("key_index", "key")
	db.Model(Member{}).AddUniqueIndex("mobile", "mobile")
	db.Model(Apostle{}).AddUniqueIndex("token_id", "token_id")
	db.Model(Apostle{}).AddIndex("status", "status")
	db.Model(Apostle{}).AddIndex("district", "district")
	db.Model(Apostle{}).AddIndex("occupational", "occupational")
	db.Model(Apostle{}).AddIndex("parents", "father", "mother")
	db.Model(Apostle{}).AddIndex("owner_status", "owner", "status")
	db.Model(ApostleTalent{}).AddUniqueIndex("token_id", "token_id")
	db.Model(ApostleTalent{}).AddIndex("apostle_id", "apostle_id")
	db.Model(ApostleTalent{}).AddIndex("element_gold", "element_gold")
	db.Model(ApostleTalent{}).AddIndex("element_wood", "element_wood")
	db.Model(ApostleTalent{}).AddIndex("element_water", "element_water")
	db.Model(ApostleTalent{}).AddIndex("element_fire", "element_fire")
	db.Model(ApostleTalent{}).AddIndex("element_soil", "element_soil")
	db.Model(ApostleTalent{}).AddIndex("life", "life")
	db.Model(ApostleTalent{}).AddIndex("mood", "mood")
	db.Model(ApostleTalent{}).AddIndex("strength", "strength")
	db.Model(ApostleTalent{}).AddIndex("agile", "agile")
	db.Model(ApostleTalent{}).AddIndex("finesse", "finesse")
	db.Model(ApostleTalent{}).AddIndex("hp", "hp")
	db.Model(ApostleTalent{}).AddIndex("intellect", "intellect")
	db.Model(ApostleTalent{}).AddIndex("lucky", "lucky")
	db.Model(ApostleTalent{}).AddIndex("potential", "potential")
	db.Model(ApostleTalent{}).AddIndex("charm", "charm")
	db.Model(AuctionApostle{}).AddUniqueIndex("create_tx", "create_tx")
	db.Model(AuctionApostle{}).AddIndex("token_id", "token_id")
	db.Model(AuctionApostle{}).AddIndex("apostle_id", "apostle_id")
	db.Model(AuctionApostle{}).AddIndex("status", "status")
	db.Model(ApostleWorkTrade{}).AddUniqueIndex("create_tx", "create_tx")
	db.Model(ApostleWorkTrade{}).AddIndex("apostle_id", "apostle_id")
	db.Model(ApostleWorkTrade{}).AddIndex("token_id", "token_id")
	db.Model(ApostleWorkTrade{}).AddIndex("status", "status")
	db.Model(ApostleFertility{}).AddUniqueIndex("create_tx", "create_tx")
	db.Model(ApostleFertility{}).AddIndex("apostle_id", "apostle_id")
	db.Model(ApostleFertility{}).AddIndex("token_id", "token_id")
	db.Model(ApostleFertility{}).AddIndex("status", "status")
	db.Model(ApostlePregnant{}).AddUniqueIndex("tx", "tx")
	db.Model(ApostleReward{}).AddUniqueIndex("tx", "tx")
	db.Model(ApostleReward{}).AddUniqueIndex("token_id", "token_id")
	db.Model(Member{}).AddUniqueIndex("tron_wallet", "tron_wallet")
	db.Model(LandApostle{}).AddIndex("land_id", "land_id")
	db.Model(LandApostle{}).AddIndex("apostle_id", "apostle_id")
	db.Model(LandApostle{}).AddUniqueIndex("apostle", "apostle_id")
	db.Model(Account{}).AddIndex("wallet", "wallet")
	db.Model(BroadcastMessage{}).AddIndex("expired_at", "expired_at")
	db.Model(AuctionHistory{}).AddIndex("auction_asset", "auction_id", "asset_type")
	db.Model(AccountVersion{}).AddIndex("account_reason_remark", "account_id", "reason", "remark")
	db.Model(TokenSwap{}).AddUniqueIndex("swap_tx", "swap_tx")
	db.Model(TokenSwap{}).AddIndex("status", "status")
	db.Model(ApostlePet{}).AddIndex("pet_type", "pet_type")
	db.Model(ApostlePet{}).AddIndex("mirror_token_id", "mirror_token_id")
	db.Model(ApostlePet{}).AddIndex("apostle_id", "apostle_id")
	db.Model(ApostlePet{}).AddIndex("district_pet_type", "district", "pet_type")
	db.Model(PetMirror{}).AddUniqueIndex("mirror_token_id", "mirror_token_id")
	db.Model(Auction{}).AddIndex("status_last_bidder", "status", "last_bidder")
	db.Model(Auction{}).AddIndex("district", "district")
	db.Model(Auction{}).AddIndex("status_seller", "status", "seller")
	db.Model(Dapp{}).AddUniqueIndex("land_status", "land_id", "status")
	db.Model(TransactionHistory{}).AddIndex("tx_action", "tx", "action")
	db.Model(TransactionHistory{}).AddIndex("address_action", "balance_address", "action")
	// db.Model(ApostleArena{}).AddUniqueIndex("create_tx", "create_tx")
	// db.Model(ApostleArena{}).AddUniqueIndex("arena_id", "chain", "arena_id")
	// db.Model(ApostleArena{}).AddUniqueIndex("current", "start_at", "chain")
	// db.Model(ApostleArenaRecord{}).AddUniqueIndex("participate_tx", "participate_tx")
	// db.Model(ApostleArenaRecord{}).AddIndex("apostle_arena_id", "apostle_arena_id")
	// db.Model(ApostleArenaHistory{}).AddIndex("address", "address")
	db.Model(EthTransaction{}).AddIndex("status", "status")
	db.Model(Drill{}).AddUniqueIndex("token_id", "token_id")
	db.Model(Drill{}).AddIndex("owner", "owner")
	db.Model(Drill{}).AddIndex("owner_spec", "owner", "formula_id")
	db.Model(LandEquip{}).AddUniqueIndex("token_id", "drill_token_id")
	db.Model(LandEquip{}).AddIndex("owner", "owner")
	db.Model(Account{}).RemoveIndex("wallet_currency")
	db.Model(Account{}).AddUniqueIndex("wallet_currency_chain", "wallet", "currency", "chain")
	db.Model(PloLeft{}).AddUniqueIndex("origin_prize_idx", "origin", "prize")
	db.Model(PloTicket{}).AddUniqueIndex("origin_pub_key_idx", "origin", "pub_key")
	db.Model(PloRaffleRecord{}).AddIndex("origin_pub_key_idx", "origin", "pub_key")
	db.Model(FarmAPR{}).AddIndex("origin_addr_idx", "addr")
	db.Model(Equipment{}).AddUniqueIndex("equipment_token_id", "equipment_token_id")
	db.Model(Equipment{}).AddIndex("apostle_token_id", "apostle_token_id")

	db.Model(ParseTxError{}).AddUniqueIndex("tx_chain_parse_func", "tx", "chain", "parse_func")

	db.Model(MemberLoginInfo{}).AddUniqueIndex("member_id__ip_ua", "member_id", "ip", "ua")

	db.Model(TransactionScan{}).AddUniqueIndex("chain_tx", "chain", "tx")
	db.Model(TransactionScan{}).AddIndex("chain_block_number", "chain", "block_number")

	db.Model(ElementRaffle{}).AddIndex("owner_chain", "owner", "chain")
	db.Model(ElementRaffle{}).AddIndex("tx_owner_chain_element", "tx", "owner", "element")

	return nil
}
