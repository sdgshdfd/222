package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditLoggerMiddleware 审计日志中间件
func AuditLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录结束时间
		duration := time.Since(start)

		// 获取用户信息
		userID, userExists := c.Get("user_id")
		userName, _ := c.Get("user_name")

		// 获取审计日志数据
		logData, exists := c.Get("audit_log")
		if !exists {
			// 没有特定审计数据时记录基本日志
			log.Printf("[AUDIT] %s %s | %d | %dms | user_id:%v",
				c.Request.Method,
				c.Request.URL.Path,
				c.Writer.Status(),
				duration.Milliseconds(),
				userID)
			return
		}

		// 转换为map
		logMap, ok := logData.(map[string]interface{})
		if !ok {
			return
		}

		// 添加通用字段
		logMap["method"] = c.Request.Method
		logMap["path"] = c.Request.URL.Path
		logMap["status"] = c.Writer.Status()
		logMap["duration"] = duration.Milliseconds()
		logMap["client_ip"] = c.ClientIP()

		if userExists {
			logMap["user_id"] = userID
			logMap["user_name"] = userName
		}

		// 结构化日志输出
		log.Printf("[AUDIT] %s %s | %d | %dms | user_id:%v | user_name:%s | action:%s | file_id:%v",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration.Milliseconds(),
			logMap["user_id"],
			logMap["user_name"],
			logMap["action"],
			logMap["file_id"])
	}
}
