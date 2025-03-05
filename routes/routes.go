package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Hedgeho9X/TeachU/controllers"
	"github.com/Hedgeho9X/TeachU/middlewares"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to TeachU API",
		})
	})

	// 创建 auth 组
	auth := r.Group("/auth")

	// 公共路由：注册 & 登录 (放在 auth 组下，但不需要 JWT 验证)
	auth.POST("/register", controllers.Register)
	auth.POST("/login", controllers.Login)

	// 需要JWT验证的路由 (在 auth 组的子组中)
	protected := auth.Group("")
	protected.Use(middlewares.JWTAuth())
	protected.GET("/profile", controllers.Profile)
	protected.POST("/resetpassword", controllers.ResetPassword)

	return r
}
