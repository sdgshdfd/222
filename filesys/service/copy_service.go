// service/copy_service.go
package service

import (
	"context"
	"filesys/dao"
	"filesys/model"
	"fmt"
	"time"
)

type copyService struct{}

var CopyService = &copyService{}

func (s *copyService) CopyFile(
	ctx context.Context,
	fileID uint,
	newParentID uint,
) (*model.File, error) {
	// 1. 获取原文件信息
	srcFile, err := dao.Q.File.WithContext(ctx).
		Where(dao.File.ID.Eq(int32(fileID))).
		First()
	if err != nil {
		return nil, fmt.Errorf("文件不存在: %w", err)
	}

	// 2. 开始事务
	tx := dao.Q.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := int32(time.Now().Unix())

	// 3. 创建新文件记录
	newFile := &model.File{
		UserID:   srcFile.UserID,
		ParentID: int32(newParentID),
		Name:     srcFile.Name,
		Type:     srcFile.Type,
		Size:     srcFile.Size,
		VerNum:   srcFile.VerNum,
		StoreKey: srcFile.StoreKey,
		Ctime:    now,
		Mtime:    now,
	}

	if err := tx.File.Create(newFile); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 4. 如果是文件，增加存储引用
	if srcFile.Type == "file" {
		if err := updateStoreRef(tx, srcFile.StoreKey, 1); err != nil {
			tx.Rollback()
			return nil, err
		}

		// 复制版本记录
		versions, err := tx.Version.
			Where(dao.Version.FileID.Eq(srcFile.ID)).
			Find()
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		for _, v := range versions {
			newVersion := &model.Version{
				FileID:   newFile.ID,
				Size:     v.Size,
				VerNum:   v.VerNum,
				StoreKey: v.StoreKey,
				Ctime:    v.Ctime,
			}
			if err := tx.Version.Create(newVersion); err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	} else { // 文件夹类型
		// 递归复制子文件
		if err := copyFolderContents(tx, srcFile.ID, newFile.ID); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 5. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return newFile, nil
}

// 递归复制文件夹内容
func copyFolderContents(tx *dao.Query, srcFolderID, destFolderID int32) error {
	// 获取所有子文件
	children, err := tx.File.Where(dao.File.ParentID.Eq(srcFolderID)).Find()
	if err != nil {
		return err
	}

	// 复制每个子文件
	for _, child := range children {
		newChild := &model.File{
			UserID:   child.UserID,
			ParentID: destFolderID,
			Name:     child.Name,
			Type:     child.Type,
			Size:     child.Size,
			VerNum:   child.VerNum,
			StoreKey: child.StoreKey,
			Ctime:    child.Ctime,
			Mtime:    child.Mtime,
		}

		if err := tx.File.Create(newChild); err != nil {
			return err
		}

		// 增加存储引用
		if child.Type == "file" {
			if err := updateStoreRef(tx, child.StoreKey, 1); err != nil {
				return err
			}

			// 复制版本记录
			versions, err := tx.Version.
				Where(dao.Version.FileID.Eq(child.ID)).
				Find()
			if err != nil {
				return err
			}

			for _, v := range versions {
				newVersion := &model.Version{
					FileID:   newChild.ID,
					Size:     v.Size,
					VerNum:   v.VerNum,
					StoreKey: v.StoreKey,
					Ctime:    v.Ctime,
				}
				if err := tx.Version.Create(newVersion); err != nil {
					return err
				}
			}
		} else { // 子文件夹
			if err := copyFolderContents(tx, child.ID, newChild.ID); err != nil {
				return err
			}
		}
	}

	return nil
}
