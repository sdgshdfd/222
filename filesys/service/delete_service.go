// service/delete_service.go
package service

import (
	"context"
	"filesys/dao"
	"fmt"
)

type deleteService struct{}

var DeleteService = &deleteService{}

func (s *deleteService) DeleteFile(ctx context.Context, fileID uint) error {
	// 1. 获取文件信息
	file, err := dao.Q.File.WithContext(ctx).
		Where(dao.Q.File.ID.Eq(int32(fileID))).
		First()
	if err != nil {
		return fmt.Errorf("文件不存在: %w", err)
	}

	// 2. 开始事务
	tx := dao.Q.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 3. 删除文件记录
	if _, err := tx.File.Where(dao.File.ID.Eq(file.ID)).Delete(); err != nil {
		tx.Rollback()
		return err
	}

	// 4. 处理文件类型
	if file.Type == "file" {
		// 删除版本记录
		if _, err := tx.Version.Where(dao.Version.FileID.Eq(file.ID)).Delete(); err != nil {
			tx.Rollback()
			return err
		}

		// 更新存储引用计数
		if err := updateStoreRef(tx, file.StoreKey, -1); err != nil {
			tx.Rollback()
			return err
		}
	} else { // 文件夹类型
		// 递归删除子文件
		if err := deleteFolderContents(tx, file.ID); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 5. 提交事务
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// 递归删除文件夹内容
func deleteFolderContents(tx *dao.Query, folderID int32) error {
	// 获取所有子文件
	children, err := tx.File.Where(dao.File.ParentID.Eq(folderID)).Find()
	if err != nil {
		return err
	}

	// 递归删除每个子文件
	for _, child := range children {
		if child.Type == "file" {
			if _, err := tx.Version.Where(dao.Version.FileID.Eq(child.ID)).Delete(); err != nil {
				return err
			}
			if err := updateStoreRef(tx, child.StoreKey, -1); err != nil {
				return err
			}
		} else {
			if err := deleteFolderContents(tx, child.ID); err != nil {
				return err
			}
		}

		if _, err := tx.File.Where(dao.File.ID.Eq(child.ID)).Delete(); err != nil {
			return err
		}
	}

	return nil
}
