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
	// 公共路由：注册 & 登录
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	// 需要JWT验证的路由
	auth := r.Group("/auth")
	auth.Use(middlewares.JWTAuth())
	auth.GET("/profile", controllers.Profile)
	auth.POST("/resetpassword", controllers.ResetPassword) // 添加重置密码路由

	return r
}
