package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware 管理员认证中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证是否为 admin 用户
		// 这里可以从会话信息中获取用户名并验证
		username := "admin" // 这里暂时写死，实际应从会话获取
		if username != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "只有管理员可以访问此接口"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// SessionAuthMiddleware 会话认证中间件
func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sid, err := c.Cookie("sid")
		if err != nil || sid == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			c.Abort()
			return
		}
		// 验证 sid 的有效性
		// 这里可以实现会话验证逻辑，例如使用数据库或缓存
		c.Next()
	}
}
