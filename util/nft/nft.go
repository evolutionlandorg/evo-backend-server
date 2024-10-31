package nft

import "context"

type NfTokenDisplay struct {
	Name           string `json:"name"`
	TokenId        string `json:"token_id"`
	ImageUrl       string `json:"image_url"`
	ImageUrlPng    string `json:"image_url_png"`
	BindApostle    bool   `json:"bind_apostle"`
	PetType        string `json:"pet_type"`
	ApostleTokenId string `json:"apostle_token_id"`
	MirrorTokenId  string `json:"mirror_token_id"`
	Amount         int    `json:"amount"`
}

type Nft interface {
	AllOwnerNft(ctx context.Context, owner string, exclude []string, page, row int, chain string, additional ...string) ([]NfTokenDisplay, int)
	NftInfo(ctx context.Context, id string) *NfTokenDisplay
	Transfer(ctx context.Context, owner, tokenId string, neg bool, chain string)
}
