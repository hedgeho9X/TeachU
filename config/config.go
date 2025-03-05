package config

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

const (
	JWTSecret = "hhhwudile.cnm!!!" // 在生产环境中应该使用环境变量
)

func ConnectDB() {
	// 修改为你自己的数据库用户名、密码、端口等
	dsn := "root:Hedgeho9@@tcp(127.0.0.1:3306)/teach_u?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	DB = db
	fmt.Println("数据库连接成功")
}
