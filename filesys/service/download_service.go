// service/download_service.go
package service

import (
	"errors"
	"filesys/dao"
	"filesys/storage"
	"os"

	"github.com/gin-gonic/gin"
)

type downloadService struct{}

var DownloadService = &downloadService{}

func (s *downloadService) Download(c *gin.Context, fileID uint) ([]byte, string, error) {
	// 1. 获取文件信息
	file, err := dao.Q.File.Where(dao.File.ID.Eq(int32(fileID))).First()
	if err != nil {
		return nil, "", err
	}

	// 2. 读取文件内容
	content, err := storage.ReadFile(file.StoreKey)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", errors.New("文件不存在或已被删除")
		}
		return nil, "", err
	}

	return content, file.Name, nil
}
