// endpoint/upload_file.go
package endpoint

import (
	"filesys/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	// 1. 获取父目录ID
	parentID, err := strconv.Atoi(c.Param("file_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的目录ID"})
		return
	}

	// 2. 获取上传文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件"})
		return
	}

	// 3. 调用服务层
	fileInfo, err := service.UploadService.Upload(c, fileHeader, uint(parentID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, fileInfo)
}
