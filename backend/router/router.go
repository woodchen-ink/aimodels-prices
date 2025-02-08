package router

import (
	"github.com/gin-gonic/gin"

	"aimodels-prices/handlers"
	"aimodels-prices/middleware"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 添加数据库中间件
	r.Use(middleware.Database())

	// 认证相关路由
	auth := r.Group("/auth")
	{
		auth.GET("/status", handlers.GetAuthStatus)
		auth.POST("/login", handlers.Login)
		auth.POST("/logout", handlers.Logout)
	}

	// 供应商相关路由
	providers := r.Group("/providers")
	{
		providers.GET("", handlers.GetProviders)
		providers.Use(middleware.RequireAuth())
		providers.Use(middleware.RequireAdmin())
		providers.POST("", handlers.CreateProvider)
		providers.PUT("/:id", handlers.UpdateProvider)
		providers.DELETE("/:id", handlers.DeleteProvider)
	}

	return r
}
