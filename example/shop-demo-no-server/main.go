package main

import (
	"fmt"
	"log"
	"shop-demo/config"
	"shop-demo/handler"
	"shop-demo/middleware"
	"shop-demo/repo"
	"shop-demo/service"

	xgdnpay "github.com/skylark8866/paysdk"
	"github.com/skylark8866/paysdk/sse"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	repository, err := repo.New(cfg.Database.Path)
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	client := xgdnpay.NewClient()
	paySSE := xgdnpay.NewPaySSE()
	defer paySSE.Shutdown()

	userService := service.NewUserService(repository, cfg.JWT.Secret)
	rechargeService := service.NewRechargeService(repository, client, paySSE)

	userHandler := handler.NewUserHandler(userService)
	rechargeHandler := handler.NewRechargeHandler(rechargeService, userHandler)
	pageHandler := handler.NewPageHandler(userService, userHandler)

	r := gin.Default()

	tmpl, err := loadTemplates()
	if err != nil {
		log.Fatal("加载模板失败:", err)
	}
	r.SetHTMLTemplate(tmpl)

	r.GET("/", pageHandler.Index)
	r.GET("/login", pageHandler.Login)
	r.GET("/register", pageHandler.Register)
	r.GET("/recharge", pageHandler.Recharge)
	r.GET("/history", rechargeHandler.History)

	api := r.Group("/api")
	{
		api.POST("/register", userHandler.Register)
		api.POST("/login", userHandler.Login)
		api.POST("/logout", userHandler.Logout)
		api.GET("/user/info", middleware.Auth(userService), userHandler.GetInfo)
		api.POST("/recharge/create", middleware.Auth(userService), rechargeHandler.CreateOrder)
		api.GET("/recharge/status", rechargeHandler.GetStatus)
		api.GET("/events/:channel", middleware.Auth(userService), paySSE.Hub().GinHandler(sse.WithConnectMessage()))
	}

	fmt.Printf("充值商城服务启动在 http://localhost:%s\n", cfg.Server.Port)
	fmt.Println("使用 Ctrl+C 停止服务")
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("服务启动失败:", err)
	}
}
