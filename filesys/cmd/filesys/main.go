package main

import (
	"filesys/dao"
	"filesys/router"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 初始化数据库连接
	db, err := gorm.Open(sqlite.Open("../../filesys.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败: ", err)
	}

	// 设置默认DAO实例
	dao.SetDefault(db)

	// 初始化路由
	r := router.InitRouter()

	// 启动服务
	log.Println("文件系统服务启动，监听端口 :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务启动失败: ", err)
	}
}
