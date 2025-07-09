package service

import (
	"context"
	"filesys/dao"
	"filesys/model"
	"fmt"
	"time"
)

type moveService struct{}

var MoveService = &moveService{}

func (s *moveService) MoveFile(
	ctx context.Context,
	fileID uint,
	newParentID uint,
) (*model.File, error) {
	// 1. 获取文件信息
	file, err := dao.Q.File.WithContext(ctx).
		Where(dao.Q.File.ID.Eq(int32(fileID))).
		First()
	if err != nil {
		return nil, fmt.Errorf("文件不存在: %w", err)
	}

	// 2. 检查目标位置是否存在同名文件
	exists, err := dao.Q.File.WithContext(ctx).
		Where(dao.Q.File.UserID.Eq(file.UserID)).
		Where(dao.Q.File.ParentID.Eq(int32(newParentID))).
		Where(dao.Q.File.Name.Eq(file.Name)).
		Exists()
	if err != nil {
		return nil, err
	}

	// 3. 自动重命名如果存在冲突
	newName := file.Name
	if exists {
		newName = generateUniqueName(file.Name, func(name string) bool {
			exists, _ := dao.Q.File.WithContext(ctx).
				Where(dao.Q.File.UserID.Eq(file.UserID)).
				Where(dao.Q.File.ParentID.Eq(int32(newParentID))).
				Where(dao.Q.File.Name.Eq(name)).
				Exists()
			return exists
		})
	}

	// 4. 开始事务
	tx := dao.Q.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 5. 更新文件位置
	updateData := map[string]interface{}{
		"parent_id": int32(newParentID),
		"name":      newName,
		"mtime":     int32(time.Now().Unix()),
	}

	if _, err := tx.File.
		Where(dao.File.ID.Eq(file.ID)).
		Updates(updateData); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 6. 如果是文件夹，更新所有子文件的路径
	if file.Type == "folder" {
		if err := updateChildPaths(tx, file.ID, newParentID); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 7. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 8. 返回更新后的文件信息
	updatedFile, err := dao.Q.File.WithContext(ctx).
		Where(dao.Q.File.ID.Eq(file.ID)).
		First()

	return updatedFile, err
}

// 递归更新子文件路径
func updateChildPaths(tx *dao.Query, folderID, newParentID uint) error {
	// 获取所有子文件
	children, err := tx.File.
		Where(dao.File.ParentID.Eq(int32(folderID))).
		Find()
	if err != nil {
		return err
	}

	// 更新每个子文件的父目录
	for _, child := range children {
		// 更新子文件的父目录为新目录的ID
		if _, err := tx.File.
			Where(dao.File.ID.Eq(child.ID)).
			Update(dao.File.ParentID, int32(newParentID)); err != nil {
			return err
		}

		// 如果是文件夹，递归更新其子文件
		if child.Type == "folder" {
			if err := updateChildPaths(tx, uint(child.ID), newParentID); err != nil {
				return err
			}
		}
	}

	return nil
}
