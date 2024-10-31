// Package routes provides ...
package routes

import (
	"net/http"

	"github.com/evolutionlandorg/evo-backend/models"

	"github.com/gin-gonic/gin"
)

type FReq struct {
	Addr string `form:"addr" binding:"required"`
}

// @Summary	Get farm APR
// @Param		addr	query	string	true	"address"
// @Tags		farm
// @Success	200	{object}	routes.GinJSON{data=map[string]string}
// @Router		/farm/apr [get]
func farmAPR() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := new(FReq)
		if err := c.ShouldBindQuery(p); err != nil {
			getReturnDataByError(c, 10001, err.Error())
			return
		}
		apr := models.RawFarmAPR(c, p.Addr)
		c.JSON(http.StatusOK, gin.H{
			"code":   0,
			"detail": "success",
			"data": map[string]string{
				"apr":     apr.APR,
				"symbol":  "%",
				"decimal": "2",
			},
		})
	}
}
