package routes

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
)

type ApiHandle struct {
	RouterGroup *gin.RouterGroup
}

func (ap *ApiHandle) StartHttpApi() {
	store := persistence.NewInMemoryStore(time.Second)
	handleCache := func(store persistence.CacheStore, expire time.Duration, handle gin.HandlerFunc) gin.HandlerFunc {
		if util.IsProduction() {
			return cache.CachePageAtomic(store, expire, handle)
		}
		return handle
	}
	api := ap.RouterGroup
	api.Use(headerMaker())

	api.POST("snapshot/vote/", Snapshot())

	// system
	api.GET("common/time", timeHandle())

	// land
	api.GET("lands", handleCache(store, time.Minute, landListHandle()))
	api.GET("land", landHandle())

	api.GET("land/rank", handleCache(store, time.Minute, landsRank()))

	api.GET("nft/metadata/:token_id", nftMetadata())

	// apostle
	api.GET("apostle/list", handleCache(store, time.Second*30, apostleListHandle()))
	api.GET("apostle/info", apostleHandle())

	api.GET("furnace/illustrated", illustrated())
	api.GET("furnace/prop", furnaceProp())
	api.GET("furnace/props", furnaceProps())

	// farm
	api.GET("farm/apr", farmAPR())

	// equipment
	api.GET("equipment/list", handleCache(store, time.Minute, equipmentList()))
	api.GET("equipment/info", handleCache(store, time.Minute, equipmentInfo()))
}

func getReturnDataByError(c *gin.Context, code int, msg ...string) {
	detail := util.QYError{Code: code}
	c.Writer.Header().Set("content-type", "application/json; charset=utf-8")
	res := gin.H{"code": code, "detail": detail.GetCode()}
	if len(msg) > 0 {
		res["message"] = strings.Join(msg, ",")
	}
	c.AbortWithStatusJSON(http.StatusOK, res)
}

func headerMaker() gin.HandlerFunc {
	return func(context *gin.Context) {
		const NetworkHeader = "EVO-NETWORK"
		chain := context.Request.Header.Get(NetworkHeader)
		var bodyEncode string
		if context.Request.Method == "POST" {
			_ = context.Request.ParseMultipartForm(10000)
			bodyEncode = context.Request.PostForm.Encode()
		}
		if chain == "" {
			chain = models.EthChain
		}
		if qChain := context.Query(NetworkHeader); qChain != "" {
			chain = qChain
		}
		headerQuery := fmt.Sprintf("cache=%x", md5.Sum([]byte(url.QueryEscape(chain+bodyEncode))))
		if context.Request.URL.RawQuery == "" {
			context.Request.URL.RawQuery += headerQuery
		} else {
			context.Request.URL.RawQuery += "&" + headerQuery
		}
		context.Set("EvoNetwork", chain)
		context.Next()
	}
}
