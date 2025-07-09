package main

import (
	"crypto/sha256"
	"encoding/hex"
	"filesys/model_def"
	"fmt"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// 清空数据库所有数据（保留表结构）
func clearAllData(db *gorm.DB) error {
	tables := []string{
		"tb_session",
		"tb_version",
		"tb_file",
		"tb_store_ref",
		"tb_user",
	}

	for _, table := range tables {
		if err := db.Exec("DELETE FROM " + table).Error; err != nil {
			return fmt.Errorf("清空表 %s 失败: %v", table, err)
		}
	}
	return nil
}

func main() {
	// 使用无 CGO 的 SQLite 驱动
	db, err := gorm.Open(sqlite.Open("../../filesys.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("打开数据库失败:", err)
	}

	// 自动建表
	models := []interface{}{
		&model_def.User{},
		&model_def.File{},
		&model_def.Version{},
		&model_def.StoreRef{},
		&model_def.Session{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Fatalf("建表失败: %v", err)
		}
	}

	// 清空现有数据
	if err := clearAllData(db); err != nil {
		log.Printf("清空数据失败: %v", err)
	}

	// 创建默认管理员用户
	adminPass := "123456"
	hash := sha256.Sum256([]byte(adminPass))
	hashStr := hex.EncodeToString(hash[:])

	adminUser := &model_def.User{
		Name:     "admin",
		Password: hashStr,
		Ctime:    time.Now().Unix(),
		Mtime:    time.Now().Unix(),
	}

	if err := db.Create(adminUser).Error; err != nil {
		log.Printf("创建管理员失败: %v", err)
	} else {
		log.Printf("管理员创建成功: 用户名=admin, 密码=%s", adminPass)
	}

	log.Println("数据库初始化完成")
}
