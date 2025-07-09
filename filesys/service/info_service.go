// service/info_service.go
package service

import (
	"context"
	"filesys/dao"
	"filesys/model"
	"fmt"
)

type infoService struct{}

var InfoService = &infoService{}

func (s *infoService) GetFileInfo(ctx context.Context, fileID uint) (*model.File, []*model.Version, error) {
	// 获取文件基本信息
	file, err := dao.Q.File.WithContext(ctx).
		Where(dao.File.ID.Eq(int32(fileID))).
		First()
	if err != nil {
		return nil, nil, fmt.Errorf("文件不存在: %w", err)
	}

	// 获取版本历史（如果是文件）
	var versions []*model.Version
	if file.Type == "file" {
		versions, err = dao.Q.Version.WithContext(ctx).
			Where(dao.Version.FileID.Eq(file.ID)).
			Order(dao.Version.VerNum.Desc()).
			Find()
		if err != nil {
			return nil, nil, err
		}
	}

	return file, versions, nil
}
