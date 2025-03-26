package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Hedgeho9X/TeachU/controllers"
	"github.com/Hedgeho9X/TeachU/middlewares"
)

// RegisterAuthRoutes 注册认证相关路由
func RegisterAuthRoutes(r *gin.Engine) {
	// 创建 auth 组
	auth := r.Group("/auth")

	auth.POST("/register", controllers.Register)
	auth.POST("/login", controllers.Login)
	auth.POST("/email2login", controllers.Email2Login)
	auth.POST("/sendcode", controllers.SendCode)
	auth.POST("/verifycode", controllers.VerifyCode)
	auth.POST("/forgetpassword", controllers.ForgetPassword)

	// 需要JWT验证的路由 (在 auth 组的子组中)
	protected := auth.Group("")
	protected.Use(middlewares.JWTAuth())
	protected.POST("/resetpassword", controllers.ResetPassword)
}
