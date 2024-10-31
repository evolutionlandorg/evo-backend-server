package cryptokitties

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/evolutionlandorg/evo-backend/util/nft"
	"math"
	"strings"
	"sync"

	"github.com/evolutionlandorg/evo-backend/util"
)

type Nft struct {
	ContractAddress   string
	AbiFile           string
	TokensIdsCacheKey string
}

type CryptoKitties struct {
	Id         uint   `json:"id"`
	Generation uint   `json:"generation"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	Bio        string `json:"bio"`
	ImageUrl   string `json:"image_url"`
	Owner      struct {
		Address string `json:"address"`
	} `json:"owner"`
	ImageUrlPng string `json:"image_url_png"`
}

type CkApiOwnerHave struct {
	Total    int             `json:"total"`
	NextPage bool            `json:"next_page"`
	Kitties  []CryptoKitties `json:"kitties"`
}

const (
	Name         = "CryptoKitties"
	infoApi      = "https://api.cryptokitties.co/kitties/%s"
	ownerHaveApi = "https://api.cryptokitties.co/v2/kitties?offset=%d&limit=20&owner_wallet_address=%s&include=other&parents=false&authenticated=true&orderBy=id&orderDirection=desc"
	cacheKey     = "CryptoKittiesName:%s:info"
)

func New() *Nft {
	tokensIdsCacheKey := Name
	return &Nft{
		ContractAddress:   util.GetContractAddress(Name),
		AbiFile:           Name,
		TokensIdsCacheKey: tokensIdsCacheKey,
	}
}

func (n *Nft) AllOwnerNft(ctx context.Context, owner string, exclude []string, page, row int, chain string, additional ...string) ([]nft.NfTokenDisplay, int) {
	var OwnerNftTokenId = func(owner string) []string {
		key := n.TokensIdsCacheKey + ":" + owner + ":" + chain
		redisKittiesId := util.SmembersCache(ctx, key)

		if len(redisKittiesId) == 0 && util.IsProduction() {
			if redisKittiesId = n.dealNftList(ctx, owner); len(redisKittiesId) > 0 {
				var b []interface{}
				for i := range redisKittiesId {
					b = append(b, redisKittiesId[i])
				}
				util.SaddArray(ctx, key, b)
				redisKittiesId = util.SmembersCache(ctx, key)
			}
		}
		return redisKittiesId
	}

	owner = strings.ToLower(owner)
	tokenIds := OwnerNftTokenId(owner)
	if len(exclude) > 0 {
		var newTokenIds []string
		for _, v := range tokenIds {
			if !util.StringInSlice(v, exclude) {
				newTokenIds = append(newTokenIds, v)
			}
		}
		tokenIds = newTokenIds
	}
	tokenIds = util.UniqueStrings(append(tokenIds, additional...))
	var cryptoKitties []nft.NfTokenDisplay
	count := len(tokenIds)
	if count <= 0 {
		return cryptoKitties, 0
	}

	maxLength := (page + 1) * row
	if maxLength > count {
		maxLength = count
	}

	tokenIds = tokenIds[page*row : maxLength]

	for _, tokenId := range tokenIds {
		if ck := n.NftInfo(ctx, tokenId); ck != nil {
			cryptoKitties = append(cryptoKitties, *ck)
		}
	}
	return cryptoKitties, count
}

func (n *Nft) NftInfo(ctx context.Context, id string) *nft.NfTokenDisplay {
	byteCk := util.GetCache(ctx, fmt.Sprintf(cacheKey, id))

	var ck nft.NfTokenDisplay
	if byteCk != nil {
		_ = json.Unmarshal(byteCk, &ck)
		return &ck
	}

	ckBytes := util.HttpGet(fmt.Sprintf(infoApi, id))
	if ckBytes == nil {
		return nil
	}

	var wck CryptoKitties
	_ = json.Unmarshal(ckBytes, &wck)
	ck = nft.NfTokenDisplay{Name: ck.Name, ImageUrlPng: wck.ImageUrlPng, ImageUrl: wck.ImageUrl, TokenId: util.IntToString(int(wck.Id)), PetType: Name}
	byteCk, _ = json.Marshal(ck)
	_ = util.SetCache(ctx, fmt.Sprintf(cacheKey, id), byteCk, 86400*7)
	return &ck
}

func (n *Nft) dealNftList(ctx context.Context, owner string) (kittiesId []string) {
	var dealOne = func(owner string, offset int) (kittiesId []string, total int) {
		ckBytes := util.HttpGet(fmt.Sprintf(ownerHaveApi, offset, owner))
		if ckBytes == nil {
			return
		}

		var cks CkApiOwnerHave
		_ = json.Unmarshal(ckBytes, &cks)

		total = cks.Total
		if total == 0 {
			return
		}

		for _, v := range cks.Kitties {
			kittiesId = append(kittiesId, util.IntToString(int(v.Id)))
			byteCk, _ := json.Marshal(nft.NfTokenDisplay{Name: v.Name, ImageUrlPng: v.ImageUrlPng, ImageUrl: v.ImageUrl, TokenId: util.IntToString(int(v.Id)), PetType: Name})
			_ = util.SetCache(ctx, fmt.Sprintf(cacheKey, util.IntToString(int(v.Id))), byteCk, 86400*7)
		}
		return
	}
	main, total := dealOne(owner, 0)
	if total == 0 {
		return
	}
	kittiesId = append(kittiesId, main...)
	page := int(math.Ceil(float64(total) / float64(20)))
	var wg sync.WaitGroup
	for i := 1; i < page; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			piece, _ := dealOne(owner, i*20)
			kittiesId = append(kittiesId, piece...)
		}(i)
	}
	wg.Wait()
	return
}

func (n *Nft) Transfer(ctx context.Context, owner, tokenId string, neg bool, chain string) {
	if util.IsProduction() {
		util.DelCache(ctx, Name+":"+owner)
	} else {
		key := n.TokensIdsCacheKey + ":" + owner + ":" + chain
		if !neg {
			util.SaddCache(ctx, key, tokenId)
			return
		}
		util.SremCache(ctx, key, tokenId)
	}
}
