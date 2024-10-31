package routes

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/gin-gonic/gin"
)

// @Summary	list apostle
// @Produce	json
// @Tags		apostle
// @Param		page			query		int		false	"page"
// @Param		row				query		int		false	"row"
// @Param		district		query		int		false	"network district, polygon:5,crab:3,eth:1,heco:4,tron:2"
// @Param		filter			query		string	false	"filter by apostle status, default is ””, in (onsell,fertility,rent,bid,unclaimed,my,unbind,fresh,sire,reward,canWorking,mine,listing,employment)"
// @Param		display			query		string	false	"Whether to filter all lands. If it is empty, only filter my lands,default is 'all'"
// @Param		order_field		query		string	false	"order field, default is 'token_index'"
// @Param		order			query		string	false	"order, default is 'asc'"
// @Param		gender			query		string	false	"gender, male, female"
// @Param		search_id		query		string	false	"search by apostle token id"
// @Param		sire_id			query		string	false	"search by sire token id"
// @Param		occupational	query		string	false	"has Guard,Saber or”"
// @Param		attribute		query		string	false	"filter by attribute"
// @Param		genesis			query		string	false	"1 or 0"
// @Success	200				{object}	routes.GinJSON{data=[]models.ApostleJson}
// @Router		/apostle/list [get]
func apostleListHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		page := util.StringToInt(c.DefaultQuery("page", "0"))
		row := util.StringToInt(c.DefaultQuery("row", "5000"))
		district := util.StringToInt(c.DefaultQuery("district", "-1"))
		filter := c.DefaultQuery("filter", "") // all
		display := c.DefaultQuery("display", "all")
		gen := util.StringToInt(c.DefaultQuery("gen", "-1"))
		orderField := c.DefaultQuery("order_field", "token_index")
		if strings.EqualFold(orderField, "time") {
			orderField = "token_index"
		}
		order := c.DefaultQuery("order", "asc")
		gender := c.Query("gender") // male，female
		searchId := c.Query("search_id")
		sireTokenId := c.Query("sire_id")
		occupational := util.ReplaceStrings(c.QueryArray("occupational"), "normal", "")
		attribute := util.StringToInt(c.DefaultQuery("attribute", "-1"))
		genesis := util.StringToInt(c.DefaultQuery("genesis", "-1"))
		chain := c.GetString("EvoNetwork")

		if orderField == "id" {
			orderField = "token_index"
		}
		query := models.ApostleQuery{Page: page, Row: row, Filter: filter, OrderField: orderField, Order: order, Display: display, Chain: chain}
		query.MultiFilter.Gen = c.QueryMap("gens")
		query.MultiFilter.Element = c.QueryArray("element")
		query.MultiFilter.Price = c.QueryMap("price")
		query.MultiFilter.Talent = c.QueryMap("talent")

		if district != -1 {
			query.WhereQuery.District = district
		}
		if gen != -1 {
			query.WhereQuery.Gen = gen
		}
		if searchId != "" {
			findNum := regexp.MustCompile("[0-9]+").FindAllString(searchId, -1)
			if len(findNum) > 0 {
				query.WhereQuery.TokenIndex = util.StringToInt(findNum[0])
			}
		}
		if gender != "" {
			query.WhereQuery.Gender = gender
		}
		if attribute != -1 {
			query.IncludeId = models.GetAttributeApostleId(c, attribute)
			query.IncludeId = append(query.IncludeId, -1)
		}
		if genesis == 1 {
			query.WhereQuery.OriginAddress = util.GetContractAddress("Gen0", chain)
		}
		memberInfo := models.AuthOwner(c)
		if memberInfo == nil && display == "my" {
			c.JSON(http.StatusOK, gin.H{"code": 0, "detail": "success", "data": []string{}, "count": 0})
			return
		}
		if memberInfo != nil && display != "all" {
			wallet := memberInfo.GetUseAddress(chain)
			switch filter {
			case "onsell":
				tokenIds, hasBidArr, _, tokenMap := models.OnsellApostleList(c, wallet, chain, nil)
				query.TokenId = tokenIds
				query.HasBid = hasBidArr
				query.Tokens = tokenMap
			case "fertility":
				query.WhereQuery.Status = "fertility"
				query.WhereQuery.OriginId = memberInfo.ID
			case "rent":
				query.WhereQuery.Status = "rent"
				query.WhereQuery.OriginId = memberInfo.ID
			case "bid":
				tokenIds, myLastBid, _ := models.BidingList(c, district, wallet, "apostle")
				query.TokenId = tokenIds
				query.MyLastBid = myLastBid
			case "unclaimed":
				tokenIds := models.UnClaimedApostleList(c, district, []string{wallet})
				query.TokenId = tokenIds
			case "my":
				query.WhereQuery.Owner = wallet
			case "unbind":
				query.WhereQuery.Owner = wallet
			case "fresh":
				query.WhereQuery.Owner = wallet
				query.WhereQuery.Status = "fresh"
			case "sire":
				query.TokenId = models.CanSiringApostle(c, wallet, sireTokenId)
			case "reward":
				query.TokenId = models.UnClaimApostleReward(c, wallet)
			case "canWorking":
				query.WhereQuery.Owner = wallet
			case "mine":
				query.WhereQuery.Owner = wallet
			case "listing":
				// onsell,fertility,rent
				query.WhereQuery.OriginId = memberInfo.ID
			case "employment":
				query.WhereQuery.Owner = wallet
			}
		} else {
			query.WhereQuery.Status = filter
		}
		if display == "all" && filter == "" && len(query.MultiFilter.Element) == 0 && len(query.MultiFilter.Gen) == 0 && len(query.MultiFilter.Price) == 0 && len(query.MultiFilter.Talent) == 0 {
			list, count := query.AllApostles(util.GetContextByGin(c), occupational)
			c.JSON(http.StatusOK, gin.H{"code": 0, "detail": "success", "data": list, "count": count})
		} else {
			list, count := query.Apostles(util.GetContextByGin(c), occupational)
			c.JSON(http.StatusOK, gin.H{"code": 0, "detail": "success", "data": list, "count": count})
		}

	}
}

// @Summary	apostle info
// @Tags		apostle
// @Param		token_id	query		string	true	"token id"
// @Success	200			{object}	routes.GinJSON{data=models.ApostleJson}
// @Router		/apostle/info [get]
func apostleHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenId := c.Query("token_id")
		if tokenId == "" {
			getReturnDataByError(c, 10001)
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "detail": "success", "data": models.GetApostleByTokenId(util.GetContextByGin(c), tokenId).AsJson(util.GetContextByGin(c))})
	}
}
