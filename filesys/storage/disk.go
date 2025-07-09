package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

// GenerateStoreKey 生成存储键（内容哈希）
func GenerateStoreKey(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

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

// ReadFile 读取文件内容
func ReadFile(storeKey string) ([]byte, error) {
	path := filepath.Join("data", storeKey[:2], storeKey)
	return os.ReadFile(path)
}

// DeleteFile 删除文件
func DeleteFile(storeKey string) error {
	path := filepath.Join("data", storeKey[:2], storeKey)
	return os.Remove(path)
}

// FileExists 检查文件是否存在
func FileExists(storeKey string) bool {
	path := filepath.Join("data", storeKey[:2], storeKey)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
