// service/upload_service.go
package service

import (
	"filesys/dao"
	"filesys/model"
	"filesys/storage"
	"io"
	"mime/multipart"
	"time"

	"github.com/gin-gonic/gin"
)

type uploadService struct{}

var UploadService = &uploadService{}

func (s *uploadService) Upload(c *gin.Context, fileHeader *multipart.FileHeader, parentID uint) (*model.File, error) {
	// 1. 读取文件内容
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// 2. 保存到磁盘
	storeKey, err := storage.SaveFile(content)
	if err != nil {
		return nil, err
	}

	// 3. 创建数据库记录
	now := int32(time.Now().Unix())
	newFile := &model.File{
		UserID:   int32(c.GetInt("user_id")),
		ParentID: int32(parentID),
		Name:     fileHeader.Filename,
		Type:     "file",
		Size:     int32(len(content)),
		VerNum:   1, // 初始版本
		StoreKey: storeKey,
		Ctime:    now,
		Mtime:    now,
	}

	// 4. 开始数据库事务
	tx := dao.Q.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 5. 创建文件记录
	if err := tx.File.Create(newFile); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 6. 创建版本记录
	version := &model.Version{
		FileID:   newFile.ID,
		Size:     int32(len(content)),
		VerNum:   1,
		StoreKey: storeKey,
		Ctime:    now,
	}

	if err := tx.Version.Create(version); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 7. 更新存储引用计数
	if err := updateStoreRef(tx, storeKey, 1); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 8. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return newFile, nil
}

// 更新存储引用计数
func updateStoreRef(tx *dao.Query, storeKey string, delta int32) error {
	ref, err := tx.StoreRef.Where(dao.StoreRef.StoreKey.Eq(storeKey)).First()
	if err != nil {
		// 不存在则创建
		return tx.StoreRef.Create(&model.StoreRef{
			StoreKey: storeKey,
			RefCount: delta,
			Ctime:    int32(time.Now().Unix()),
			Mtime:    int32(time.Now().Unix()),
		})
	}

	// 更新计数
	newCount := ref.RefCount + delta
	if newCount <= 0 {
		// 删除引用记录
		if _, err := tx.StoreRef.Delete(ref); err != nil {
			return err
		}
		// 物理删除文件
		return storage.DeleteFile(storeKey)
	}

	_, err = tx.StoreRef.Where(dao.StoreRef.ID.Eq(ref.ID)).
		Update(dao.StoreRef.RefCount, newCount)
	return err
}
