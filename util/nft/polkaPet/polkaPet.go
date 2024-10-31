package polkaPet

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/evolutionlandorg/evo-backend/util/nft"
	"math/big"
	"strings"

	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"
)

type Nft struct {
	ContractAddress   string
	TokensIdsCacheKey string
}

const (
	Name     = "PolkaPets"
	infoApi  = "https://api.ppw.digital/api/item/%s"
	cacheKey = "PolkaPet:%s:info"
)

type PolkaPet struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

var tokenIds = []string{"2", "11", "20"}

func New() *Nft {
	return &Nft{ContractAddress: util.GetContractAddress("PolkaPet"), TokensIdsCacheKey: Name}
}

func (n *Nft) AllOwnerNft(ctx context.Context, owner string, _ []string, _, _ int, _ string, _ ...string) ([]nft.NfTokenDisplay, int) {

	var dealNftList = func(owner string) map[string]int {
		sg := storage.New("Eth")
		var tokenIdsBigInt []*big.Int
		for _, v := range tokenIds {
			tokenIdsBigInt = append(tokenIdsBigInt, big.NewInt(int64(util.StringToInt(v))))
		}
		amountSlice := sg.BalanceOfBatch(n.ContractAddress, []string{owner, owner, owner}, tokenIdsBigInt)
		if amountSlice == nil {
			return nil
		}
		list := make(map[string]int)
		for index, v := range tokenIds {
			list[v] = int(amountSlice[index])
		}
		return list
	}
	owner = strings.ToLower(owner)
	tokenIds := dealNftList(owner)
	var list []nft.NfTokenDisplay
	count := len(tokenIds)
	if count <= 0 {
		return list, 0
	}

	for tokenId, amount := range tokenIds {
		if ck := n.NftInfo(ctx, tokenId); ck != nil {
			ck.Amount = amount
			list = append(list, *ck)
		}
	}
	return list, count
}

func (n *Nft) NftInfo(ctx context.Context, id string) *nft.NfTokenDisplay {
	raw := util.GetCache(ctx, fmt.Sprintf(cacheKey, id))

	var pet nft.NfTokenDisplay
	if raw != nil {
		_ = json.Unmarshal(raw, &pet)
		return &pet
	}

	raw = util.HttpGet(fmt.Sprintf(infoApi, id))
	if raw == nil {
		return nil
	}

	var info PolkaPet
	_ = json.Unmarshal(raw, &info)

	display := nft.NfTokenDisplay{Name: info.Name, ImageUrlPng: info.Image, ImageUrl: info.Image, TokenId: id, PetType: Name}
	raw, _ = json.Marshal(display)

	_ = util.SetCache(ctx, fmt.Sprintf(cacheKey, id), raw, 86400*7)
	return &display
}

func (n *Nft) Transfer(_ context.Context, _, _ string, _ bool, _ string) {

}
