package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/win30221/core/basic"
	"github.com/win30221/core/config"
	commonDelivery "github.com/win30221/core/http/delivery"
	"github.com/win30221/core/http/middleware"
	"github.com/win30221/core/storage"
	_ "github.com/win30221/point/docs"
	"github.com/win30221/point/service/delivery"
	"github.com/win30221/point/service/repository"
	"github.com/win30221/point/service/usecase"
)

const (
	ServerName = "point"
)

// build version
// go build 時使用 -ldflags 傳入
var (
	VERSION   string
	COMMIT    string
	BUILDTIME string
)

// @title point 模組
// @description 積分錢包
// @securityDefinitions.apikey Systoken
// @in header
// @name Systoken

func main() {
	basic.Version = VERSION
	basic.Commit = COMMIT
	basic.BuildTime = BUILDTIME

	basic.Init(ServerName)

	// 建立 HTTP server
	r := gin.New()

	// 設定 http server Middleware
	r.Use(gin.Recovery())
	r.Use(middleware.Log()...)

	// 設定基礎路由
	_, privateGroup := commonDelivery.SetBasicRouter(r)

	// ttl
	balanceTTL, _ := config.GetSeconds("/service/point/balance_ttl", true)
	pointLogsTTL, _ := config.GetSeconds("/service/point/point_logs_ttl", true)

	mysqlM := storage.GetMysqlDB("/storage/mysql/master/celluloid-picket")
	mysqlS := storage.GetMysqlDB("/storage/mysql/slave/celluloid-picket")
	redis := storage.GetRedis("/storage/redis/celluloid-picket", ServerName)

	// repo
	mysqlRepo := repository.NewMysqlRepo(mysqlM, mysqlS)
	redisRepo := repository.NewRedisRepo(redis, balanceTTL, pointLogsTTL)

	// use case
	useCase := usecase.NewUseCase(
		mysqlRepo,
		redisRepo,
	)
	delivery.NewDelivery(privateGroup, useCase)

	r.Run(fmt.Sprintf("%s:%s", basic.Host, basic.Port))
}
