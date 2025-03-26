package routes

import (
	"github.com/Hedgeho9X/TeachU/controllers"
	"github.com/Hedgeho9X/TeachU/middlewares"
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
	exam.DELETE("/delete/:id", controllers.DeleteExam)
	exam.GET("/list", controllers.ListExam)
	exam.POST("/upload-score", controllers.UploadScore)
}
