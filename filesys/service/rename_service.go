// service/rename_service.go
package service

import (
	"context"
	"filesys/dao"
	"filesys/model"
	"fmt"
	"time"
)

type renameService struct{}

var RenameService = &renameService{}

func (s *renameService) RenameFile(
	ctx context.Context,
	fileID uint,
	newName string,
) (*model.File, error) {
	// 1. 获取文件信息
	file, err := dao.Q.File.WithContext(ctx).
		Where(dao.File.ID.Eq(int32(fileID))).
		First()
	if err != nil {
		return nil, fmt.Errorf("文件不存在: %w", err)
	}

	// 2. 检查新名称是否冲突
	exists, err := dao.Q.File.WithContext(ctx).
		Where(dao.File.UserID.Eq(file.UserID)).
		Where(dao.File.ParentID.Eq(file.ParentID)).
		Where(dao.File.Name.Eq(newName)).
		Exists()
	if err != nil {
		return nil, err
	}

	// 3. 自动重命名如果存在冲突
	finalName := newName
	if exists {
		finalName = generateUniqueName(newName, func(name string) bool {
			exists, _ := dao.Q.File.WithContext(ctx).
				Where(dao.File.UserID.Eq(file.UserID)).
				Where(dao.File.ParentID.Eq(file.ParentID)).
				Where(dao.File.Name.Eq(name)).
				Exists()
			return exists
		})
	}

	// 4. 更新文件名
	updateData := map[string]interface{}{
		"name":  finalName,
		"mtime": int32(time.Now().Unix()),
	}

	if _, err := dao.Q.File.
		Where(dao.File.ID.Eq(file.ID)).
		Updates(updateData); err != nil {
		return nil, err
	}

	// 5. 返回更新后的文件信息
	updatedFile, err := dao.Q.File.WithContext(ctx).
		Where(dao.File.ID.Eq(file.ID)).
		First()

	return updatedFile, err
}
