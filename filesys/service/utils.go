package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"filesys/dao"
	"filesys/model"
	"filesys/storage"

	"gorm.io/gorm"
)

// 更新存储引用计数（事务内使用）
func updateStoreRef(tx *dao.Query, storeKey string, delta int32) error {
	ref, err := tx.StoreRef.Where(dao.StoreRef.StoreKey.Eq(storeKey)).First()
	if err != nil {
		// 不存在则创建
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.StoreRef.Create(&model.StoreRef{
				StoreKey: storeKey,
				RefCount: delta,
				Ctime:    int32(time.Now().Unix()),
				Mtime:    int32(time.Now().Unix()),
			})
		}
		return err
	}

	// 更新计数
	newCount := ref.RefCount + delta
	if newCount <= 0 {
		// 删除引用记录
		if _, err := tx.StoreRef.Delete(ref); err != nil {
			return err
		}
		// 物理删除文件
		if storage.FileExists(storeKey) {
			return storage.DeleteFile(storeKey)
		}
		return nil
	}

	_, err = tx.StoreRef.Where(dao.StoreRef.ID.Eq(ref.ID)).
		Update(dao.StoreRef.RefCount, newCount)
	return err
}

// 生成唯一文件名
func generateUniqueName(original string, existsFunc func(string) bool) string {
	base, ext := splitFilename(original)

	for i := 1; ; i++ {
		newName := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if !existsFunc(newName) {
			return newName
		}
	}
}

// 分离文件名和扩展名
func splitFilename(filename string) (string, string) {
	ext := filepath.Ext(filename)
	base := filename[:len(filename)-len(ext)]
	return base, ext
}

// 计算密码哈希
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
