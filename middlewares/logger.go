package middlewares

import (
	"fmt"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/gin-gonic/gin"

	"time"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		var params string
		if c.Request.Method == "GET" {
			params = c.Request.URL.RawQuery
		} else {
			if data := c.Request.PostForm.Encode(); len(data) < 2048 {
				params = data
			}
		}
		msg := fmt.Sprintf("%v -- %v -- %v -- %s -- %s -- %s",
			c.Writer.Status(),
			latency,
			c.Request.Method,
			c.Request.URL.Path,
			params,
			c.Request.Header.Get("EVO-NETWORK"))
		log.Info(msg)

	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer util.Recover(fmt.Sprintf("%s recover", c.Request.URL.String()))
		c.Next()
	}
}
