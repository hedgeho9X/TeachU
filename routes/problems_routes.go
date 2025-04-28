package routes

import (
	"github.com/Hedgeho9X/TeachU/internal/controllers"
	"github.com/Hedgeho9X/TeachU/internal/middlewares"
	"github.com/gin-gonic/gin"
)

// RegisterProblemsRoutes 注册题目管理相关路由
func RegisterProblemsRoutes(r *gin.Engine) {
	problems := r.Group("/problems")
	problems.Use(middlewares.JWTAuth())
	problems.POST("/create", controllers.CreatProblem)
}
