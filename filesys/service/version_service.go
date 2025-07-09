// service/version_service.go
package service

import (
	"context"
	"errors"
	"filesys/dao"
	"filesys/storage"
	"fmt"
	"os"
	"path/filepath"
)

type versionService struct{}

var VersionService = &versionService{}

func (s *versionService) DownloadVersion(
	ctx context.Context,
	fileID uint,
	versionNum int,
) ([]byte, string, error) {
	// 1. 获取版本信息
	version, err := dao.Q.Version.WithContext(ctx).
		Where(dao.Version.FileID.Eq(int32(fileID))).
		Where(dao.Version.VerNum.Eq(int32(versionNum))).
		First()
	if err != nil {
		return nil, "", fmt.Errorf("版本不存在: %w", err)
	}

	// 2. 获取文件信息（用于文件名）
	file, err := dao.Q.File.WithContext(ctx).
		Where(dao.File.ID.Eq(int32(fileID))).
		First()
	if err != nil {
		return nil, "", fmt.Errorf("文件不存在: %w", err)
	}

	// 3. 生成版本文件名
	ext := filepath.Ext(file.Name)
	base := file.Name[:len(file.Name)-len(ext)]
	versionedName := fmt.Sprintf("%s (v%d)%s", base, versionNum, ext)

	// 4. 读取文件内容
	content, err := storage.ReadFile(version.StoreKey)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", errors.New("文件内容已丢失")
		}
		return nil, "", err
	}

	return content, versionedName, nil
}
