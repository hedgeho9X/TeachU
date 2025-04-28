package main

import (
	"fmt"
	"log"

	"github.com/Hedgeho9X/TeachU/internal/config"
	"github.com/Hedgeho9X/TeachU/routes"
)

func main() {
	// 1. 连接数据库
	config.ConnectDB()

	// 2. 初始化 Redis 连接
	config.InitRedis()

	// 3. 设置路由
	r := routes.SetupRouter()

	// 4. 启动服务
	fmt.Println("Server running 8080!")
	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
