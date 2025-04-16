package main

import (
	"fmt"
	"log"

	"github.com/Hedgeho9X/TeachU/internal/config"
	"github.com/Hedgeho9X/TeachU/internal/models"
	"github.com/Hedgeho9X/TeachU/routes"
)

func main() {
	// 1. 连接数据库
	config.ConnectDB()

	// 2. 初始化 Redis 连接
	config.InitRedis()

	// 3. 数据表迁移
	if err := config.DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("数据迁移失败:", err)
	}

	// 4. 设置路由
	r := routes.SetupRouter()

	// 5. 启动服务
	fmt.Println("Server running on http://localhost:8080")
	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
