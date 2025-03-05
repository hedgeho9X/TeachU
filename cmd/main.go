package main

import (
	"fmt"
	"log"

	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
	"github.com/Hedgeho9X/TeachU/routes"
)

func main() {
	// 1. 连接数据库
	config.ConnectDB()

	// 2. 数据表迁移
	err := config.DB.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("数据迁移失败:", err)
	}

	// 3. 设置路由
	r := routes.SetupRouter()

	// 4. 启动服务
	fmt.Println("Server running on http://localhost:8080")
	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatal("服务器启动失败:", err)
	}

}
