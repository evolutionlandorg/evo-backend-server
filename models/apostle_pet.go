package models

import (
	"context"
	"errors"
	"github.com/evolutionlandorg/evo-backend/util/nft"
	"github.com/evolutionlandorg/evo-backend/util/nft/cryptokitties"
	"github.com/evolutionlandorg/evo-backend/util/nft/polkaPet"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/jinzhu/gorm"
)

type ApostlePet struct {
	Id             string `json:"id"`
	ApostleId      uint   `json:"apostle_id"`
	ApostleTokenId string `json:"apostle_token_id"`
	PetType        string `json:"pet_type"`
	Chain          string `json:"chain"`
	TokenId        string `json:"token_id"`
	Name           string `json:"name"`
	ImageUrl       string `json:"image_url"`
	MirrorTokenId  string `json:"mirror_token_id"`
	District       int    `json:"district"`
	Owner          string `json:"owner"`
}

type ApostlePetJson struct {
	ApostleId      uint   `json:"apostle_id"`
	PetType        string `json:"pet_type"`
	TokenId        string `json:"token_id"`
	Name           string `json:"name"`
	ImageUrl       string `json:"image_url"`
	ApostleTokenId string `json:"apostle_token_id"`
	MirrorTokenId  string `json:"mirror_token_id"`
}

const (
	CryptoKittiesPetType = "CryptoKitties"
	PolkaPetType         = "PolkaPets"
)

func (ap *ApostlePet) New(txn *util.GormDB) error {
	result := txn.Create(&ap)
	return result.Error
}

type ApostlePeteQuery struct {
	WhereQuery struct {
		Owner   string
		Chain   string
		PetType string
	}
	Row    int
	Page   int
	Filter string
}

type PetMap struct {
	ApostleTokenId string `json:"apostle_token_id"`
	MirrorTokenId  string `json:"mirror_token_id"`
}

type OwnerNfToken struct {
	nft.NfTokenDisplay
	Pet2Apostle []PetMap `json:"pet_2_apostle"`
}

func (apq *ApostlePeteQuery) GetOwnerNfToken(ctx context.Context) (nfts []OwnerNfToken, count int) {
	var (
		bindTokenIds, mirrorTokenIds, exclude []string
	)
	var list []nft.NfTokenDisplay
	var ori nft.Nft
	switch apq.WhereQuery.PetType {
	case CryptoKittiesPetType:
		ori = cryptokitties.New()
	case PolkaPetType:
		ori = polkaPet.New()
	default:
		return
	}
	pet2Apostle := make(map[string]string)
	pet2Mirror := make(map[string]string)
	mirrorApostle := make(map[string][]PetMap)

	bindToken := getApostleBindPet(ctx, apq.WhereQuery.Owner, apq.WhereQuery.PetType, apq.WhereQuery.Chain)
	for _, v := range bindToken {

		mirrorApostle[v.TokenId] = append(mirrorApostle[v.TokenId], PetMap{MirrorTokenId: v.MirrorTokenId, ApostleTokenId: v.ApostleTokenId})

		pet2Mirror[v.TokenId] = v.MirrorTokenId
		mirrorTokenIds = append(mirrorTokenIds, v.TokenId)
		if v.ApostleId > 0 {
			bindTokenIds = append(bindTokenIds, v.TokenId)
			pet2Apostle[v.TokenId] = v.ApostleTokenId
		}
	}
	additional := mirrorTokenIds
	switch apq.Filter {
	case "unbind": // 未绑定
		exclude = bindTokenIds
	}
	list, count = ori.AllOwnerNft(ctx, apq.WhereQuery.Owner, exclude, apq.Page, apq.Row, apq.WhereQuery.Chain, additional...)
	for _, v := range list {
		if util.StringInSlice(v.TokenId, mirrorTokenIds) {
			v.MirrorTokenId = pet2Mirror[v.TokenId]
		}
		if util.StringInSlice(v.TokenId, bindTokenIds) {
			v.BindApostle = true
			v.ApostleTokenId = pet2Apostle[v.TokenId]
		}
		nftToken := OwnerNfToken{NfTokenDisplay: v, Pet2Apostle: mirrorApostle[v.TokenId]}
		if len(nftToken.Pet2Apostle) == 0 && v.Amount == 0 && strings.EqualFold(apq.WhereQuery.PetType, PolkaPetType) {
			count -= 1
			continue
		}
		nfts = append(nfts, nftToken)
	}
	return nfts, count
}

func (ap *Apostle) updateApostlePetCount(ctx context.Context, neg bool) error {
	doQuery := gorm.Expr("bind_pet_count + ?", 1)
	if neg {
		doQuery = gorm.Expr("bind_pet_count - ?", 1)
	}
	query := util.WithContextDb(ctx).Model(&ap).Update(map[string]interface{}{"bind_pet_count": doQuery})
	return query.Error
}

func (ec *EthTransactionCallback) PetBaseCallback(ctx context.Context) (err error) {
	if getTransactionDeal(ctx, ec.Tx, "petBase") != nil {
		return errors.New("tx exist")
	}
	txn := util.DbBegin(ctx)
	defer txn.DbRollback()
	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("petBase", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			case services.AbiEncodingMethod("Tied(uint256,uint256,uint256,bool,address,address)"):
				dataSlice := util.LogAnalysis(log.Data)
				mirrorTokenId := dataSlice[1]
				apostleTokenId := dataSlice[0]
				talent := dataSlice[2]
				apostle := GetApostleByTokenId(ctx, apostleTokenId)
				if apostle == nil {
					return errors.New("apostle token id error")
				}
				from := util.AddHex(dataSlice[5][24:64], chain)
				txn.Model(ApostlePet{}).Where("mirror_token_id = ?", mirrorTokenId).UpdateColumn(map[string]interface{}{
					"apostle_id": apostle.ID, "apostle_token_id": apostle.TokenId,
				})
				if err := apostle.RefreshTalent(txn, talent); err != nil {
					txn.DbRollback()
					return err
				}
				if err := apostle.updateApostlePetCount(ctx, false); err != nil {
					txn.DbRollback()
					return err
				}
				th := TransactionHistory{Tx: ec.Tx, Chain: chain, BalanceAddress: from, Action: TransactionHistoryApostleBindPet, TokenId: apostleTokenId}
				_ = th.New(txn)
			case services.AbiEncodingMethod("UnTied(uint256,uint256,uint256,bool,address,address)"):
				dataSlice := util.LogAnalysis(log.Data)
				mirrorTokenId := dataSlice[1]
				apostleTokenId := dataSlice[0]
				talent := dataSlice[2]
				from := util.AddHex(dataSlice[5][24:64], chain)
				apostle := GetApostleByTokenId(ctx, apostleTokenId)
				if apostle == nil {
					return errors.New("apostle token id error")
				}
				txn.Table("apostle_pets").Where("mirror_token_id = ?", mirrorTokenId).UpdateColumn(map[string]interface{}{"apostle_id": 0, "apostle_token_id": ""})
				if err := apostle.RefreshTalent(txn, talent); err != nil {
					txn.DbRollback()
					return err
				}
				if err := apostle.updateApostlePetCount(ctx, true); err != nil {
					txn.DbRollback()
					return err
				}
				th := TransactionHistory{Tx: ec.Tx, Chain: chain, BalanceAddress: from, Action: TransactionHistoryApostleUnbindPet, TokenId: apostleTokenId}
				_ = th.New(txn)
			}
		}
	}

	if err := ec.NewUniqueTransaction(txn, "petBase"); err != nil {
		return err
	}

	txn.DbCommit()
	return txn.Error
}

func getApostleBindPet(ctx context.Context, owner, petType, chain string) []ApostlePetJson {
	db := util.WithContextDb(ctx)
	var apj []ApostlePetJson
	db.Table("apostle_pets").Where("owner = ? and pet_type= ? and chain= ?", owner, petType, chain).Scan(&apj)
	return apj
}

func getNfTokenByTokenId(ctx context.Context, tokenId, chain, petTokenAddress string) *nft.NfTokenDisplay {
	switch petTokenAddress {
	case util.GetContractAddress("cryptoKitties", chain):
		pet := cryptokitties.New()
		return pet.NftInfo(ctx, tokenId)
	case util.GetContractAddress("polkaPet", chain):
		pet := polkaPet.New()
		return pet.NftInfo(ctx, tokenId)
	}
	return nil
}

func getApostlePetsReversalKey(ctx context.Context, ids []uint) map[uint]ApostlePetJson {
	if len(ids) < 1 {
		return nil
	}
	db := util.WithContextDb(ctx)
	var pj []ApostlePetJson
	query := db.Table("apostle_pets").Where("apostle_id in (?)", ids).Scan(&pj)
	if query.Error != nil || query == nil || query.RecordNotFound() {
		return nil
	}
	results := make(map[uint]ApostlePetJson)
	for _, v := range pj {
		results[v.ApostleId] = v
	}
	return results
}
