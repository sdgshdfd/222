package middleware

import (
	"filesys/dao"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sid, err := c.Cookie("sid")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		// 验证 sid 是否存在
		session, err := dao.Q.Session.WithContext(c.Request.Context()).
			Where(dao.Session.SessionID.Eq(sid)).First()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效sid"})
			c.Abort()
			return
		}
		user, err := dao.Q.User.WithContext(c.Request.Context()).
			Where(dao.User.ID.Eq(session.UserID)).First()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			c.Abort()
			return
		}

		c.Set("user_id", int(user.ID))
		c.Set("user_name", user.Name)
		c.Next()
	}
}
