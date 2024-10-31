package routes

import (
	"net/http"
	"time"

	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"

	"github.com/gin-gonic/gin"
	_ "github.com/gin-gonic/gin/render"
)

type GinJSON struct {
	Data    interface{} `json:"data"`
	Code    int         `json:"code"`
	Detail  string      `json:"detail"`
	Message string      `json:"message"`
}

// @Summary	Time server time
// @Tags		common
// @Produce	json
// @Success	200	{number}	number	"server time"
// @Router		/common/time [post]
func timeHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":   0,
			"detail": "success",
			"data":   time.Now().Unix(),
		})
	}
}

// @Summary	Get Apostle,Land,Drill,Material,Equipment,MirrorKitty NFT metadata
// @Tags		common
// @Produce	json
// @Param		token_id	path		string						true	"token id"
// @Success	200			{object}	models.NftMetaData			"nft metadata"
// @Failure	404			{object}	routes.GinJSON{data=nil}	"not found"
// @Router		/common/nft/metadata/{token_id} [get]
func nftMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenId := c.Param("token_id")
		metaData, err := models.GetNftMetaData(util.GetContextByGin(c), tokenId)
		if err != nil {
			log.Debug("get nft meta data failed %s %s", err, tokenId)
			c.JSON(http.StatusNotFound, nil)
			return
		}
		c.JSON(http.StatusOK, metaData)
	}
}
