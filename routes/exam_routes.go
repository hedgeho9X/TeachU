package routes

import (
	"github.com/Hedgeho9X/TeachU/internal/controllers"
	"github.com/Hedgeho9X/TeachU/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterExamRoutes(r *gin.Engine) {
	r.Use(func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Next()
	})
	exam := r.Group("/exam")
	exam.Use(middlewares.JWTAuth())
	exam.POST("/create", controllers.CreateExam)
	exam.DELETE("/:id", controllers.DeleteExam)
	exam.GET("/list", controllers.ListExam)
	score := r.Group("/exam/score")
	score.POST("/upload", controllers.UploadScore)
	score.GET("/list", controllers.ListScore)
}
