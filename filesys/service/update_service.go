// service/update_service.go
package service

import (
	"context"
	"filesys/dao"
	"filesys/model"
	"filesys/storage"
	"io"
	"mime/multipart"
	"time"
)

type updateService struct{}

var UpdateService = &updateService{}

func (s *updateService) UpdateFile(
	ctx context.Context,
	fileHeader *multipart.FileHeader,
	fileID uint,
) (*model.File, error) {
	// 1. 获取原文件信息
	oldFile, err := dao.Q.File.WithContext(ctx).
		Where(dao.File.ID.Eq(int32(fileID))).
		First()
	if err != nil {
		return nil, err
	}

	// 2. 读取新文件内容
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// 3. 保存新文件
	newStoreKey, err := storage.SaveFile(content)
	if err != nil {
		return nil, err
	}

	// 4. 开始事务
	tx := dao.Q.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := int32(time.Now().Unix())

	// 5. 创建新版本
	newVersion := &model.Version{
		FileID:   oldFile.ID,
		Size:     int32(len(content)),
		VerNum:   oldFile.VerNum + 1,
		StoreKey: newStoreKey,
		Ctime:    now,
	}

	if err := tx.Version.Create(newVersion); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 6. 更新文件记录
	updateData := map[string]interface{}{
		"size":      int32(len(content)),
		"ver_num":   oldFile.VerNum + 1,
		"store_key": newStoreKey,
		"mtime":     now,
	}

	if _, err := tx.File.Where(dao.File.ID.Eq(oldFile.ID)).
		Updates(updateData); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 7. 更新存储引用计数
	if err := updateStoreRef(tx, newStoreKey, 1); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 减少旧文件引用
	if err := updateStoreRef(tx, oldFile.StoreKey, -1); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 8. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 9. 返回更新后的文件信息
	updatedFile, err := tx.File.WithContext(ctx).
		Where(dao.File.ID.Eq(oldFile.ID)).
		First()

	return updatedFile, err
}
