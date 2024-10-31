package models

import (
	"context"
	"errors"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
)

type PetMirror struct {
	gorm.Model
	PetType       string `json:"pet_type"`
	Chain         string `json:"chain"`
	TokenId       string `json:"token_id"`
	MirrorTokenId string `json:"mirror_token_id"`
	Owner         string `json:"owner"`
}

func (pm *PetMirror) New(txn *util.GormDB) error {
	result := txn.Create(&pm)
	if result.Error == nil {
		refreshOpenSeaMetadata(pm.MirrorTokenId)
	}
	return result.Error
}

func (ec *EthTransactionCallback) NftBridgeCallback(ctx context.Context) (err error) {
	if getTransactionDeal(ctx, ec.Tx, "nftBridge") != nil {
		return errors.New("tx exist")
	}
	txn := util.DbBegin(ctx)
	defer txn.DbRollback()
	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("nftBridge", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			// SwapIn(address originContract, uint256 originTokenId, uint256 mirrorTokenId, address owner)
			case services.AbiEncodingMethod("SwapIn(address,uint256,uint256,address)"):
				dataSlice := util.LogAnalysis(log.Data)
				tokenId := util.U256(dataSlice[1]).String()
				mirrorTokenId := dataSlice[2]
				pet := getNfTokenByTokenId(ctx, tokenId, chain, util.AddHex(dataSlice[0][24:64], chain))
				if pet == nil {
					return errors.New("nft not found")
				}
				owner := util.AddHex(dataSlice[3][24:64], chain)
				var record ApostlePet
				if q := txn.Model(ApostlePet{}).Where("mirror_token_id = ?", dataSlice[2]).First(&record); q.RecordNotFound() {
					ap := ApostlePet{Chain: chain, PetType: pet.PetType, TokenId: tokenId, MirrorTokenId: mirrorTokenId, Name: pet.Name, ImageUrl: pet.ImageUrlPng, Owner: owner}
					if err := ap.New(txn); err != nil {
						txn.DbRollback()
						return err
					}
				} else {
					txn.Model(ApostlePet{}).Where("mirror_token_id = ?", dataSlice[2]).Update("owner", owner)
				}
				// SwapOut(address originContract, uint256 originTokenId, uint256 mirrorTokenId, address owner)
			case services.AbiEncodingMethod("SwapOut(address,uint256,uint256,address)"):
				dataSlice := util.LogAnalysis(log.Data)
				txn.Model(ApostlePet{}).Where("mirror_token_id = ?", dataSlice[2]).Update("owner", util.GetContractAddress("nftBridge", chain))
			}
		}
	}

	if err != nil {
		return err
	}

	if err := ec.NewUniqueTransaction(txn, "nftBridge"); err != nil {
		return err
	}

	txn.DbCommit()
	return txn.Error
}

func getPetTokenIdByMirror(ctx context.Context, mirrorTokenId string) *PetMirror {
	db := util.WithContextDb(ctx)
	var pm PetMirror
	query := db.Model(PetMirror{}).First(&pm, PetMirror{MirrorTokenId: mirrorTokenId})
	if query.Error != nil || query.RecordNotFound() {
		return nil
	}
	return &pm
}
