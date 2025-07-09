// service/list_service.go
package service

import (
	"context"
	"filesys/dao"
	"filesys/model"
	"sort"
)

type listService struct{}

var ListService = &listService{}

func (s *listService) ListFiles(ctx context.Context, parentID uint) ([]*model.File, error) {
	// 获取当前用户ID
	userID := ctx.Value("user_id").(int32)

	// 查询目录下的文件
	files, err := dao.Q.File.WithContext(ctx).
		Where(dao.File.UserID.Eq(userID)).
		Where(dao.File.ParentID.Eq(int32(parentID))).
		Find()
	if err != nil {
		return nil, err
	}

	// 按文件夹优先排序
	sort.Slice(files, func(i, j int) bool {
		if files[i].Type == "folder" && files[j].Type != "folder" {
			return true
		}
		if files[i].Type != "folder" && files[j].Type == "folder" {
			return false
		}
		return files[i].Name < files[j].Name
	})

	return files, nil
}
