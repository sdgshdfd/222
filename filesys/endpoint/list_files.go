package endpoint

import (
	"filesys/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListFiles 获取文件列表端点
// @Summary 获取目录下的文件列表
// @Description 列出指定目录下的所有文件和文件夹
// @Tags 文件操作
// @Accept json
// @Produce json
// @Param file_id path int true "目录ID"
// @Security ApiKeyAuth
// @Success 200 {array} model.File "文件列表"
// @Failure 400 {object} gin.H "{"error": "无效的目录ID"}"
// @Failure 403 {object} gin.H "{"error": "无权限访问该目录"}"
// @Failure 500 {object} gin.H "{"error": "获取文件列表失败"}"
// @Router /api/file/{file_id}/list [get]
func ListFiles(c *gin.Context) {
	// 1. 获取目录ID
	fileID, err := GetFileIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的目录ID"})
		return
	}

	// 2. 记录操作日志
	LogAction(c, "list_files", fileID)

	// 3. 调用服务层
	files, err := service.ListService.ListFiles(c.Request.Context(), fileID)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	// 4. 返回结果
	c.JSON(http.StatusOK, files)
}
