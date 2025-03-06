package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"aimodels-prices/config"
	"aimodels-prices/database"
	"aimodels-prices/handlers"
	"aimodels-prices/middleware"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.DB.Close()

	// 设置gin模式
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 注入数据库
	r.Use(func(c *gin.Context) {
		c.Set("db", database.DB)
		c.Next()
	})

	// CORS中间件
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API路由组
	api := r.Group("/api")
	{
		// 价格相关路由
		prices := api.Group("/prices")
		{
			prices.GET("", handlers.GetPrices)
			prices.GET("/rates", handlers.GetPriceRates)
			prices.POST("", middleware.AuthRequired(), handlers.CreatePrice)
			prices.PUT("/:id", middleware.AuthRequired(), handlers.UpdatePrice)
			prices.DELETE("/:id", middleware.AuthRequired(), handlers.DeletePrice)
			prices.PUT("/:id/status", middleware.AuthRequired(), middleware.AdminRequired(), handlers.UpdatePriceStatus)
			prices.PUT("/approve-all", middleware.AuthRequired(), middleware.AdminRequired(), handlers.ApproveAllPrices)
		}

		// 模型厂商相关路由
		providers := api.Group("/providers")
		{
			providers.GET("", handlers.GetProviders)
			providers.POST("", middleware.AuthRequired(), handlers.CreateProvider)
			providers.PUT("/:id", middleware.AuthRequired(), middleware.AdminRequired(), handlers.UpdateProvider)
			providers.DELETE("/:id", middleware.AuthRequired(), middleware.AdminRequired(), handlers.DeleteProvider)
		}

		// 认证相关路由
		auth := api.Group("/auth")
		{
			auth.GET("/status", handlers.GetAuthStatus)
			auth.POST("/login", handlers.Login)
			auth.POST("/logout", handlers.Logout)
			auth.GET("/user", handlers.GetUser)
			auth.GET("/callback", handlers.AuthCallback)
		}

		// 模型类型相关路由
		modelTypes := api.Group("/model-types")
		{
			modelTypes.GET("", handlers.GetModelTypes)
			modelTypes.POST("", middleware.AuthRequired(), handlers.CreateModelType)
		}
	}

	// 启动服务器
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
