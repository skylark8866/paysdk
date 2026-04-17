package main

import (
	"context"
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

	if cfg.Payment.AppID == "" || cfg.Payment.AppSecret == "" {
		log.Fatal("请配置 payment.app_id 和 payment.app_secret")
	}

	repository, err := repo.New(cfg.Database.Path)
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	client := xgdnpay.NewClient(
		cfg.Payment.AppID,
		cfg.Payment.AppSecret,
		xgdnpay.WithBaseURL(cfg.Payment.BaseURL),
	)

	sseHub := sse.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sseHub.Run(ctx)

	userService := service.NewUserService(repository, cfg.JWT.Secret)
	rechargeService := service.NewRechargeService(repository, client, cfg)
	rechargeService.SetSSEHub(sseHub)

	userHandler := handler.NewUserHandler(userService)
	rechargeHandler := handler.NewRechargeHandler(rechargeService, userHandler)
	pageHandler := handler.NewPageHandler(userService, userHandler)

	r := gin.Default()

	// 使用 embed.FS 加载模板
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
		api.GET("/events/:channel", middleware.Auth(userService), sseHub.GinHandler(sse.WithConnectMessage()))
	}

	r.POST("/api/callback/pay", func(c *gin.Context) {
		var req xgdnpay.NotifyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.String(400, "参数错误")
			return
		}

		if err := xgdnpay.VerifyNotify(&req, cfg.Payment.AppSecret, 300); err != nil {
			c.String(401, "签名验证失败")
			return
		}

		if err := rechargeHandler.HandleCallback(&req); err != nil {
			c.String(500, err.Error())
			return
		}

		c.String(200, "SUCCESS")
	})

	fmt.Printf("充值商城服务启动在 http://localhost:%s\n", cfg.Server.Port)
	fmt.Println("使用 Ctrl+C 停止服务")
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("服务启动失败:", err)
	}
}
