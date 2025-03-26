package routes

import (
	"github.com/Hedgeho9X/TeachU/controllers"
	"github.com/Hedgeho9X/TeachU/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterProblemsRoutes(r *gin.Engine) {
	problems := r.Group("/problems")
	problems.Use(middlewares.JWTAuth())
	problems.POST("/create", controllers.CreatProblem)
}
