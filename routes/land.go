package routes

import (
	"fmt"
	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/util"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

// @Summary	Get land list
// @Tags		land
// @Produce	json
// @Param		page		query		int		false	"page"
// @Param		row			query		int		false	"row"
// @Param		display		query		string	false	"Whether to filter all lands. If it is empty, only filter my lands,default is 'all'"
// @Param		district	query		int		false	"network district, polygon:5,crab:3,eth:1,heco:4,tron:2. default is 1"
// @Param		filter		query		string	false	"filter by land status, default is ””, in (unclaimed,bid,onsale,my,other,fresh,mine,gold_rush,availableDrill,genesis,secondhand,plo)"
// @Param		order_field	query		string	false	"order field, default is 'token_index'"
// @Param		order		query		string	false	"order, default is 'desc'"
// @Param		search_id	query		string	false	"search by token index"
// @Param		address		query		string	false	"search by owner address"
// @Success	200			{object}	routes.GinJSON{data=[]models.LandJson}
// @Router		/lands [get]
func landListHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		chain := c.GetString("EvoNetwork")

		page := util.StringToInt(c.DefaultQuery("page", "0"))
		row := util.StringToInt(c.DefaultQuery("row", "2025"))
		display := c.DefaultQuery("display", "all")
		district := util.StringToInt(c.DefaultQuery("district", "1"))
		filter := c.DefaultQuery("filter", "") // my,unclaimed,bid,onsale
		orderField := c.DefaultQuery("order_field", "token_index")
		order := c.DefaultQuery("order", "desc")
		searchId := c.Query("search_id")
		address := c.Query("address")
		query := models.LandQuery{Page: page, Row: row, Filter: filter, Network: chain, Order: order, OrderField: orderField}
		query.MultiFilter.Flag = c.QueryArray("flag")
		query.MultiFilter.Element = c.QueryArray("element")
		query.MultiFilter.Price = c.QueryMap("price")

		// search by token index
		if searchId != "" {
			findNum := regexp.MustCompile("[0-9]+").FindAllString(searchId, -1)
			if len(findNum) > 0 && util.StringToInt(findNum[0]) > 0 {
				query.WhereInterface = append(query.WhereInterface, fmt.Sprintf("token_id = '%s'", models.GenerateLandTokenId(chain, util.StringToInt(findNum[0]))))
			}
		}

		query.WhereQuery.District = district
		query.WhereQuery.Owner = address
		var (
			list  interface{}
			count int
		)
		if memberInfo := models.AuthOwner(c, true); memberInfo != nil && display != "all" {
			wallet := memberInfo.GetUseAddress(chain)
			switch filter {
			case "unclaimed": // 待领取地块
				tokenIds := models.MyAuctionLandList(util.GetContextByGin(c), district, []string{wallet})
				query.TokenId = tokenIds
			case "bid": // 已出价地块
				tokenIds, myLastBid, priceMap := models.BidingList(util.GetContextByGin(c), district, wallet, "land", true)
				query.TokenId = tokenIds
				query.PriceMap = priceMap
				query.MyLastBid = myLastBid
			case "onsale": // 出售中地块
				tokenIds, hasBidArr, priceMap, _, tokenMap := models.OnsellLandList(util.GetContextByGin(c), district, fmt.Sprintf("seller = '%s' and status = '%s'", wallet, "going"))
				query.TokenId = tokenIds
				query.PriceMap = priceMap
				query.HasBid = hasBidArr
				query.TokenMap = tokenMap
			case "my":
				query.WhereQuery.Owner = wallet
			case "other":
				query.WhereInterface = append(query.WhereInterface, fmt.Sprintf("owner != '%s'", wallet))
			case "fresh":
				query.WhereQuery.Owner = wallet
				query.WhereQuery.Status = "fresh"
			case "mine":
				query.WhereQuery.Owner = wallet
				tokenIds, hasBidArr, _, _, tokenMap := models.OnsellLandList(util.GetContextByGin(c), district, fmt.Sprintf("seller = '%s' and status = '%s'", wallet, "going"))
				query.TokenId = tokenIds
				query.HasBid = hasBidArr
				query.TokenMap = tokenMap
			case "gold_rush":
				query.WhereQuery.Owner = wallet
			case "availableDrill": // 地块上有使徒 && 钻头不满
				digLandIds := models.DigLands(util.GetContextByGin(c))
				if len(digLandIds) == 0 {
					c.JSON(http.StatusOK, gin.H{"code": 0, "detail": "success", "data": list, "count": 0})
					return
				}
				query.WhereInterface = append(query.WhereInterface, fmt.Sprintf("id in (%s)", util.StringsJoinQuot(digLandIds)))
				query.WhereInterface = append(query.WhereInterface, fmt.Sprintf("token_id not in (%s)",
					util.StringsJoinQuot(models.FullyLoadedLandId(util.GetContextByGin(c)))))
			}
			query.PendingTrans = models.CurrentPendingLand(util.GetContextByGin(c), wallet)
			var result *[]models.LandJson
			result, count = query.LandList(util.GetContextByGin(c))

			if filter == "unclaimed" && result != nil {
				for index := range *result {
					(*result)[index].Status = models.AuctionClaimed
				}
			}
			c.JSON(http.StatusOK, gin.H{"code": 0, "detail": "success", "data": result, "count": count})
			return
		}

		switch filter {
		case "onsale":
			tokenIds, hasBidArr, priceMap, _, tokenMap := models.OnsellLandList(util.GetContextByGin(c), district, fmt.Sprintf("status = '%s'", "going"))
			query.TokenId = tokenIds
			query.PriceMap = priceMap
			query.TokenMap = tokenMap
			query.HasBid = hasBidArr
			list, count = query.LandList(util.GetContextByGin(c))
		case "genesis":
			tokenIds, hasBidArr, priceMap, startAtMap, tokenMap := models.OnsellLandList(util.GetContextByGin(c), district,
				fmt.Sprintf("seller = '%s' and status = '%s'", util.GetContractAddress("genesisHolder", chain), "going"),
			)
			query.TokenId = tokenIds
			query.HasBid = hasBidArr
			query.PriceMap = priceMap
			query.TokenMap = tokenMap
			query.AuctionStartAtMap = startAtMap
			list, count = query.LandList(util.GetContextByGin(c))
		case "secondhand":
			tokenIds, hasBidArr, priceMap, _, tokenMap := models.OnsellLandList(util.GetContextByGin(c), district,
				fmt.Sprintf("seller != '%s' and status = '%s'", util.GetContractAddress("genesisHolder", chain), "going"),
			)
			query.TokenId = tokenIds
			query.HasBid = hasBidArr
			query.TokenMap = tokenMap
			query.PriceMap = priceMap
			list, count = query.LandList(util.GetContextByGin(c))
		case "plo":
			query.WhereInterface = append(query.WhereInterface, fmt.Sprintf("token_index in (%s)", util.StringsJoinQuot(util.Evo.GRLandId[models.GetChainByDistrict(district)])))
			list, count = query.LandList(util.GetContextByGin(c))
		default:
			_, _, priceMap, _, tokenMap := models.OnsellLandList(util.GetContextByGin(c), district, fmt.Sprintf("status = '%s'", "going"))
			query.PriceMap = priceMap
			query.TokenMap = tokenMap
			list, count = query.AllLands(util.GetContextByGin(c))
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "detail": "success", "data": list, "count": count})
	}
}

// @Summary	Get land by token_id
// @Tags		land
// @Produce	json
// @Param		token_id	query		string	true	"token_id"
// @Success	200			{object}	routes.GinJSON{data=models.LandDetailJson}
// @Router		/land [get]
func landHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenId := c.Query("token_id")
		if tokenId == "" {
			getReturnDataByError(c, 10001)
			return
		}
		memberInfo := models.AuthOwner(c, true)
		c.JSON(http.StatusOK, gin.H{"code": 0, "detail": "success", "data": models.GetLandByTokenId(util.GetContextByGin(c), tokenId).AsJson(util.GetContextByGin(c), memberInfo)})
	}
}

// @Summary	List lands rank
// @Tags		land
// @Produce	json
// @Success	200	{object}	routes.GinJSON{data=[]models.LandRank}
// @Router		/land/rank [get]
func landsRank() gin.HandlerFunc {
	return func(c *gin.Context) {
		chain := c.GetString("EvoNetwork")
		district := models.GetDistrictByChain(chain)
		c.JSON(http.StatusOK, gin.H{"code": 0, "detail": "success", "data": models.LandsRankList(util.GetContextByGin(c), district, chain)})
	}
}
