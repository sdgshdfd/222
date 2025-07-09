package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

// SaveFile 保存文件到磁盘
func SaveFile(content []byte) (string, error) {
	storeKey := GenerateStoreKey(content)
	path := filepath.Join("data", storeKey[:2], storeKey)

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}

	// 检查文件是否已存在
	if _, err := os.Stat(path); err == nil {
		return storeKey, nil // 文件已存在
	}

	// 保存文件
	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", err
	}

	return storeKey, nil
}

// GenerateStoreKey 生成文件存储键
func GenerateStoreKey(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// GetFilePath 获取文件路径
func GetFilePath(storeKey string) string {
	return filepath.Join("data", storeKey[:2], storeKey)
}
