package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/evolutionlandorg/evo-backend/commands"
	"github.com/evolutionlandorg/evo-backend/config"
	"github.com/evolutionlandorg/evo-backend/daemons"
	_ "github.com/evolutionlandorg/evo-backend/docs"
	"github.com/evolutionlandorg/evo-backend/middlewares"
	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/routes"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/urfave/cli"
	gintrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
)

var (
	ctx, cancel = context.WithCancel(context.TODO())
)

// @title Evo Backend Server
// @version         1.0

// @contact.name   API Support
// @contact.url    https://github.com/orgs/evolutionlandorg/discussions
// @license.name  MIT
// @host       backend.evolution.land
// @BasePath  /api
func main() {
	Init()
	util.Panic(setupApp().Run(os.Args))
}

func Init() {
	log.InitLog(log.Options{
		Level:         log.StrLevel2zAPlEVEL(util.GetEnv("LOG_LEVEL", "DEBUG")),
		UseDebugPanic: !util.IsProduction(),
	})

	config.InitApplication()
	util.Panic(util.InitMysql(log.NewGormLog()))
	util.Panic(models.MigrationDbTable())
	util.Panic(util.InitRedis())
	util.Panic(util.InitWorkers())
}

func setupApp() *cli.App {
	return &cli.App{
		Name:  "EVOLUTION LAND",
		Usage: "Evolution.land Backend",
		Action: func(c *cli.Context) error {
			server := &http.Server{
				Addr:    util.GetEnv("PORT", ":2333"),
				Handler: setupRouter(),
			}
			if !cast.ToBool(util.GetEnv("DISABLE_DAEMONS", "false")) {
				daemons.Start(ctx)
			}

			go func() {
				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Fatal("listen error: %s", err)
				}
			}()

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			<-quit
			cancel()
			_ = server.Shutdown(context.TODO())
			time.Sleep(time.Second * 5)
			return nil
		},
		Version:  "1.0",
		Commands: commands.SubAction,
	}
}

func setupRouter() (server *gin.Engine) {
	server = gin.New()
	server.Use(middlewares.Recovery(),
		gintrace.Middleware("EVO-BACKEND", gintrace.WithAnalytics(true)),
		middlewares.CORS(),
		middlewares.Logger())

	server.MaxMultipartMemory = 3 << 20
	server.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 10000,
			"msg":  "not found",
		})
	})
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	server.GET("/apostle/:genes", func(ctx *gin.Context) {
		apostlePictureFilePath := filepath.Join(util.ApostlePictureDir, ctx.Param("genes"))
		if _, err := os.Stat(apostlePictureFilePath); err == nil {
			ctx.File(apostlePictureFilePath)
			return
		}
		u, err := url.Parse(util.ApostlePictureServer)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "svg server url error"})
			log.Fatal("svg server url error: %s", err)
			return
		}
		u.Path = strings.ReplaceAll(ctx.Param("genes"), ".png", ".svg")
		resp, err := http.Get(u.String())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "svg server error"})
			log.Fatal("svg server error: %s", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "svg server error"})
			log.Fatal("svg server error: %s", resp.Status)
			return
		}
		body := bytes.NewBuffer(nil)
		_, err = io.Copy(body, resp.Body)
		if err != nil || body.Len() == 0 {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "svg server error"})
			log.Fatal("svg server error: %s", err)
			return
		}
		if err := os.WriteFile(filepath.Join(util.ApostlePictureDir, u.Path), body.Bytes(), 0644); err != nil {
			log.Warn("write file error: %s", err)
		}
		ctx.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), body, nil)
	})
	api := routes.ApiHandle{RouterGroup: server.Group("/api")}
	api.StartHttpApi()
	return
}
