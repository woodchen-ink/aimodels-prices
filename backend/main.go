package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"aimodels-prices/config"
	"aimodels-prices/cron"
	"aimodels-prices/database"
	"aimodels-prices/handlers"
	one_hub_handlers "aimodels-prices/handlers/one_hub"
	initTasks "aimodels-prices/init"
	"aimodels-prices/middleware"
	"aimodels-prices/seo"
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

	// 运行初始化任务
	initTasks.RunInitTasks()

	// 设置gin模式
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化并启动定时任务
	cron.Init()
	defer cron.StopCronJobs()

	r := gin.Default()

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

			prices.GET("/rates", one_hub_handlers.GetPriceRates) //one_hub 价格倍率, 旧接口

			prices.POST("", middleware.AuthRequired(), handlers.CreatePrice)
			prices.PUT("/:id", middleware.AuthRequired(), handlers.UpdatePrice)
			prices.DELETE("/:id", middleware.AuthRequired(), handlers.DeletePrice)
			// 审核价格需要t4或admin权限
			prices.PUT("/:id/status", middleware.AuthRequired(), middleware.RequireModerator(), handlers.UpdatePriceStatus)
			prices.PUT("/approve-all", middleware.AuthRequired(), middleware.RequireModerator(), handlers.ApproveAllPrices)
		}

		//one_hub 路由
		one_hub := api.Group("/one_hub")
		{
			one_hub.GET("/rates", one_hub_handlers.GetPriceRates)
			one_hub.GET("/official-rates", one_hub_handlers.GetOfficialPriceRates)
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
	}

	// 静态文件服务 - 支持 SPA
	staticDir := "./frontend"
	if _, err := os.Stat(staticDir); err == nil {
		// 处理静态文件请求
		r.NoRoute(func(c *gin.Context) {
			// 如果是 API 请求，返回 404
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
				return
			}

			// 检查文件是否存在
			path := filepath.Join(staticDir, c.Request.URL.Path)
			fileInfo, err := os.Stat(path)

			// 如果文件存在且不是目录，直接返回文件
			if err == nil && !fileInfo.IsDir() {
				c.File(path)
				return
			}

			// 如果文件不存在或是目录，返回 index.html (SPA 支持)
			// 使用 SEO 模块根据路径动态替换 meta 标签
			seo.RenderIndex(c, staticDir)
		})

		log.Printf("静态文件服务已启用，目录: %s", staticDir)
	} else {
		log.Printf("警告: 静态文件目录不存在 (%s)，仅提供 API 服务", staticDir)
	}

	// 启动服务器
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("服务器启动在端口 %s", cfg.ServerPort)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
