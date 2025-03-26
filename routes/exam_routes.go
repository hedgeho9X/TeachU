package routes

import (
	"github.com/Hedgeho9X/TeachU/controllers"
	"github.com/Hedgeho9X/TeachU/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterExamRoutes(r *gin.Engine) {
	exam := r.Group("/exam")
	exam.Use(middlewares.JWTAuth())
	exam.POST("/create", controllers.CreateExam)
	exam.DELETE("/delete/:id", controllers.DeleteExam)
	exam.GET("/list", controllers.ListExam)
}
