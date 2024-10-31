package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/services/storage"
	"github.com/evolutionlandorg/evo-backend/util"

	"github.com/gin-gonic/gin"
)

type furnacePropsParams struct {
	Row       int    `form:"row" binding:"required"`
	Page      int    `form:"page"`
	FormulaId int    `form:"formula_id"`
	Order     string `form:"order" binding:"omitempty,oneof=desc asc"`
	Filter    string `form:"filter" binding:"omitempty,oneof=fresh working"`
}

// @Summary	List furnace props
// @Tags		furnace
// @Params		row query true "row"
// @Params		page query true "page"
// @Params		formula_id query true "formula_id"
// @Params		order query true "order"
// @Params		filter query true "filter"
// @Success	200	{object}	routes.GinJSON{data=[]models.Drill}
// @Router		/furnace/props [get]
func furnaceProps() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := new(furnacePropsParams)
		chain := c.GetString("EvoNetwork")
		if err := c.ShouldBindQuery(p); err != nil {
			getReturnDataByError(c, 10001, err.Error())
			return
		}
		memberInfo := models.AuthOwner(c, true)
		if memberInfo == nil {
			getReturnDataByError(c, 99999)
			return
		}
		wallet := memberInfo.GetUseAddress(chain)
		if wallet == "" {
			getReturnDataByError(c, 10000)
			return
		}

		opt := models.ListOpt{Page: p.Page, Row: p.Row, Order: p.Order, Filter: p.Filter, Chain: chain}
		switch p.Filter {
		case "fresh":
			opt.WhereQuery = append(opt.WhereQuery, fmt.Sprintf("owner = '%s'", wallet))
		case "working":
			opt.WhereQuery = append(opt.WhereQuery, fmt.Sprintf("token_id in ('%s')", strings.Join(memberInfo.EqDrill(util.GetContextByGin(c), wallet), "','")))
		default:
			opt.WhereQuery = append(opt.WhereQuery, fmt.Sprintf("owner = '%s' or token_id in ('%s')", wallet, strings.Join(memberInfo.EqDrill(util.GetContextByGin(c), wallet), "','")))
		}
		opt.WhereQuery = append(opt.WhereQuery, fmt.Sprintf("chain = '%s'", chain))
		if p.FormulaId > 0 {
			opt.WhereQuery = append(opt.WhereQuery, fmt.Sprintf("formula_id = '%d'", p.FormulaId))
			if p.FormulaId != 256 {
				opt.Display = "ignore" // ignore dego
			}
		}
		list, count := models.Drills(util.GetContextByGin(c), opt, wallet, models.DrillMultiFilter{
			Class:   c.QueryArray("class"),
			Grade:   c.QueryArray("grade"),
			Element: c.QueryArray("element"),
		})

		var tokenIds []string
		for _, drill := range list {
			tokenIds = append(tokenIds, drill.TokenId)
		}

		// fill land equip
		if p.Filter != "fresh" {
			if eq := models.LandEquipByTokenId(util.GetContextByGin(c), tokenIds); eq != nil {
				for index, drill := range list {
					if value, ok := eq[drill.TokenId]; ok {
						list[index].LandEquip = &value
					}
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"code":   0,
			"detail": "success",
			"data":   list,
			"count":  count,
		})
	}
}

// @Summary	Furnace Prop
// @Tags		furnace
// @Success	200			{object}	routes.GinJSON{data=models.Drill}
// @Param		token_id	query		string	true	"token id"
// @Router		/furnace/prop [get]
func furnaceProp() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := new(struct {
			TokenId string `form:"token_id" binding:"required"`
		})
		chain := c.GetString("EvoNetwork")
		if err := c.ShouldBindQuery(p); err != nil {
			getReturnDataByError(c, 10001, err.Error())
			return
		}

		drill := models.GetDrillsByTokenId(util.GetContextByGin(c), p.TokenId)
		if drill == nil { // dego
			sg := storage.New(chain)
			if owner := sg.DEGOOwnerOf(p.TokenId); owner != "" {
				drill = &models.Drill{TokenId: p.TokenId, Class: 0, Grade: 1, Owner: owner, FormulaId: 256}
			}
		}
		if drill != nil {
			if eq := models.LandEquipByTokenId(c, []string{p.TokenId}); eq != nil {
				if value, ok := eq[p.TokenId]; ok {
					drill.LandEquip = &value
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"code":   0,
			"detail": "success",
			"data":   drill,
		})
	}
}

// illustrated furnace illustrated
//
//	@Tags		furnace
//	@Success	200	{object}	routes.GinJSON{data=[]util.Formula}
//	@Router		/furnace/illustrated [get]
func illustrated() gin.HandlerFunc {
	return func(c *gin.Context) {
		chain := c.GetString("EvoNetwork")
		c.JSON(http.StatusOK, gin.H{
			"code":   0,
			"detail": "success",
			"data":   models.Formulas(c, chain),
		})
	}
}
