package models

import (
	"context"
	"github.com/evolutionlandorg/evo-backend/util/nft/cryptokitties"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util"
)

var nftTransferEvent = services.AbiEncodingMethod("Transfer(address,address,uint256)")

func (ec *EthTransactionCallback) CryptoKittiesCallback(ctx context.Context) (err error) {
	chain := ec.Receipt.ChainSource
	for _, log := range ec.Receipt.Logs {
		if len(log.Topics) != 0 && strings.EqualFold(util.AddHex(log.Address, chain), util.GetContractAddress("cryptoKitties", chain)) {
			eventName := util.AddHex(log.Topics[0])
			switch eventName {
			case nftTransferEvent:
				var tokenId, from, to string
				logSlice := util.LogAnalysis(log.Data)
				tokenId = util.U256(logSlice[2]).String()
				from = util.AddHex(util.TrimHex(logSlice[0])[24:64], chain)
				to = util.AddHex(util.TrimHex(logSlice[1])[24:64], chain)
				n := cryptokitties.New()
				if GetMemberByAddress(ctx, from, chain) != nil {
					n.Transfer(ctx, from, tokenId, true, chain)
				}
				if GetMemberByAddress(ctx, to, chain) != nil {
					n.Transfer(ctx, to, tokenId, false, chain)
				}
			}
		}
	}
	return err
}

// func (ec *EthTransactionCallback) BlockchainCutiesCallback() (err error) {
// 	chain := ec.Receipt.ChainSource
// 	for _, log := range ec.Receipt.Logs {
// 		if len(log.Topics) != 0 && util.AddHex(log.Address, chain) == util.GetContractAddress("blockchainCuties", chain) {
// 			eventName := util.AddHex(log.Topics[0])
// 			switch eventName {
// 			case nftTransferEvent:
// 				var tokenId, from, to string
// 				tokenId = util.U256(log.Data).String()
// 				from = util.AddHex(util.TrimHex(log.Topics[1])[24:64], chain)
// 				to = util.AddHex(util.TrimHex(log.Topics[2])[24:64], chain)
// 				n := blockchaincuties.New(chain)
// 				if GetMemberByAddress(from, chain) != nil {
// 					n.Transfer(from, tokenId, true)
// 				}
// 				if GetMemberByAddress(to, chain) != nil {
// 					n.Transfer(to, tokenId, false)
// 				}
// 			}
// 		}
// 	}
// 	return err
// }
