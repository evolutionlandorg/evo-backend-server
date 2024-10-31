package models

import (
	"context"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
)

type TransactionScan struct {
	gorm.Model
	Tx             string `json:"tx" gorm:"varchar(84)"`
	Chain          string `json:"chain" gorm:"varchar(24)"`
	BlockNumber    int64  `json:"block_number"`
	BlockTimestamp int64  `json:"block_timestamp"`
	Logs           string `json:"logs" gorm:"type:json"`
}

func (t *TransactionScan) New(ctx context.Context) {
	util.WithContextDb(ctx).Create(t)
}

func (ec *EthTransactionCallback) ObjectOwnershipCallback(ctx context.Context) error {
	var (
		err     error
		tokenId string
	)
	txn := util.DbBegin(ctx)
	defer txn.DbRollback()
	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && util.AddHex(log.Topics[0]) == services.AbiEncodingMethod("Transfer(address,address,uint256)") && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("objectOwnership", chain)) {
			to := util.AddHex(util.TrimHex(log.Topics[2])[24:64], chain)
			if len(log.Topics) < 4 {
				tokenId = util.TrimHex(log.Data)
			} else {
				tokenId = util.TrimHex(log.Topics[3])
			}
			_ = GetOrCreateMemberByAddress(ctx, to, chain)
			switch getAssetTypeByTokenId(tokenId) {
			case AssetLand:
				err = updateLandOwner(ctx, txn, tokenId, to)
			case AssetMirrorKitty:
				txn.Model(ApostlePet{}).Where("mirror_token_id=?", tokenId).Update("owner", to)
			case AssetApostle:
				err = changeApostleOwner(txn, tokenId, to, chain)
			case AssetDrill, AssetItem:
				if drill := GetDrillsByTokenId(ctx, tokenId); drill != nil {
					err = drill.Transfer(ctx, to, tokenId)
				}
			case AssetEquipment:
				if eq := GetEquipment(ctx, tokenId); eq != nil {
					err = eq.Transfer(ctx, to, tokenId)
				}
			}
			if err != nil {
				return err
			}
		}
	}
	txn.DbCommit()
	return nil
}
