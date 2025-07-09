package endpoint

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetFileIDParam 从路径获取文件ID
func GetFileIDParam(c *gin.Context) (uint, error) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(fileID), nil
}

// GetVersionParam 从路径获取版本号
func GetVersionParam(c *gin.Context) (int, error) {
	verStr := c.Param("ver_num")
	verNum, err := strconv.Atoi(verStr)
	if err != nil || verNum < 1 {
		return 0, err
	}
	return verNum, nil
}

// LogAction 记录操作审计日志
func LogAction(c *gin.Context, action string, fileID uint) {
	userID, exists := c.Get("user_id")
	if !exists {
		return
	}

	userName, _ := c.Get("user_name")

	logData := map[string]interface{}{
		"action":    action,
		"file_id":   fileID,
		"user_id":   userID,
		"user_name": userName,
	}

	c.Set("audit_log", logData)
}

// HandleServiceError 统一处理服务层错误
func HandleServiceError(c *gin.Context, err error) {
	statusCode := http.StatusInternalServerError
	errorMsg := "操作失败: " + err.Error()

	switch {
	case strings.Contains(err.Error(), "不存在"):
		statusCode = http.StatusNotFound
	case strings.Contains(err.Error(), "无权限"):
		statusCode = http.StatusForbidden
	case strings.Contains(err.Error(), "无效的"):
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, gin.H{"error": errorMsg})
}
