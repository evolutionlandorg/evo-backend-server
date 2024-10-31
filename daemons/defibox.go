// Package daemons provides ...
package daemons

import (
	"context"
	"strconv"
	"strings"

	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/services"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/shopspring/decimal"
)

func UploadProjectData(ctx context.Context, chain string) {
	data := make(map[string]string)
	price, err := services.GetRingPrice()
	if err == nil {
		ring := "ring"
		data["tvl"] = getTVL(ctx, chain, ring, price)
		data["trans_amount_24h"] = getTransAmount24H(ctx, chain, ring, price)
		data["trans_amount_7d"] = getTransAmount7D(ctx, chain, ring, price)
	}
	data["chain"] = strings.ToUpper(chain)
	data["address_count"] = strconv.FormatUint(models.GetMemberCount(ctx, chain), 10)
	data["address_active_24h"] = strconv.FormatUint(models.GetMemberActive24H(ctx, chain), 10)
	data["address_active_7d"] = strconv.FormatUint(models.GetMemberActive7D(ctx, chain), 10)
	data["trans_count_24h"] = strconv.FormatUint(models.GetTransCount24H(ctx, chain), 10)
	data["trans_count_7d"] = strconv.FormatUint(models.GetTransCount7D(ctx, chain), 10)
	res := services.ReportDefiBoxData("/dgg/open/report/project/data", data)
	log.Info("/dgg/open/report/project/data response=%s", res)
}

func getTransAmount24H(ctx context.Context, chain, ring string, price decimal.Decimal) string {
	amount := models.GetTransAmount24H(ctx, chain, ring)
	return amount.Mul(price).String()
}

func getTransAmount7D(ctx context.Context, chain, ring string, price decimal.Decimal) string {
	amount := models.GetTransAmount7D(ctx, chain, ring)
	return amount.Mul(price).String()
}

func getTVL(ctx context.Context, chain, ring string, price decimal.Decimal) string {
	tvl := models.GetTVL(ctx, chain, ring)
	return tvl.Mul(price).String()
}
