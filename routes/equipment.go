package routes

import (
	"fmt"
	"net/http"

	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/gin-gonic/gin"
)

// @Summary	Get equipment list
// @Tags		equipment
// @Param		row		query		int										true	"row"
// @Param		page	query		int										true	"page"
// @Param		object	query		string									true	"object"
// @Param		order	query		string									true	"order"
// @Success	200		{object}	routes.GinJSON{data=[]models.Equipment}	"ok"
// @Router		/equipment/list [get]
func equipmentList() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := new(struct {
			Row    int    `form:"row" binding:"required"`
			Page   int    `form:"page"`
			Object string `form:"object" binding:"omitempty,oneof=Sword Shield" enums:"[Sword,Shield]"`
			Order  string `form:"order" binding:"omitempty,oneof=desc asc" enums:"[desc,asc]"`
		})
		chain := c.GetString("EvoNetwork")
		if err := c.ShouldBindQuery(p); err != nil {
			getReturnDataByError(c, 10001, err.Error())
			return
		}

		opt := models.ListOpt{Page: p.Page, Row: p.Row, Order: p.Order, Chain: chain, OrderField: "rarity"}
		opt.WhereQuery = []interface{}{fmt.Sprintf("chain = '%s'", chain)}
		if p.Object != "" {
			opt.WhereQuery = append(opt.WhereQuery, fmt.Sprintf("object = '%s'", p.Object))
		}
		if memberInfo := models.AuthOwner(c, true); memberInfo != nil {
			wallet := memberInfo.GetUseAddress(chain)
			if wallet == "" {
				getReturnDataByError(c, 10035)
				return
			}
			opt.WhereQuery = append(opt.WhereQuery, fmt.Sprintf("owner = '%s' or origin_owner = '%s'", wallet, wallet))
		}

		list, count := models.EquipmentList(c, opt)
		c.JSON(http.StatusOK, gin.H{"code": 0, "detail": "success", "data": list, "count": count})
	}
}

// @Summary	Get equipment info
// @Tags		equipment
// @Param		token_id	query		string										true	"token_id"
// @Success	200			{object}	routes.GinJSON{data=models.EquipmentJson}	"ok"
// @Router		/equipment/info [get]
func equipmentInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := new(struct {
			TokenId string `form:"token_id" binding:"required"`
		})
		if err := c.ShouldBindQuery(p); err != nil {
			getReturnDataByError(c, 10001, err.Error())
			return
		}

		eq := models.GetEquipment(c, p.TokenId)
		if eq == nil {
			getReturnDataByError(c, 10404)
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "detail": "success", "data": eq.AsJson(util.GetContextByGin(c))})
	}
}
