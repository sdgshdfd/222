// endpoint/download_file.go
package endpoint

import (
	"filesys/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func DownloadFile(c *gin.Context) {
	// 1. 获取文件ID
	fileID, err := strconv.Atoi(c.Param("file_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	// 2. 调用服务层
	content, filename, err := service.DownloadService.Download(c, uint(fileID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. 设置下载头
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", content)
}
