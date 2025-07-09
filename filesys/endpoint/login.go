package endpoint

import (
	"filesys/dao"
	"filesys/model"
	"filesys/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	uid, err := service.LoginService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	// 设置会话 cookie
	// 生成uuid作为会话ID
	sessionId := uuid.New().String()
	// 将会话ID存储到数据库中
	dao.Q.Session.WithContext(c.Request.Context()).Create(&model.Session{
		UserID:    uid,
		SessionID: sessionId,
		Ctime:     int32(time.Now().Unix()),                    // 创建时间
		Etime:     int32(time.Now().Add(1 * time.Hour).Unix()), // 过期时间，1小时后
	})
	c.SetCookie("sid", sessionId, 3600, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "登录成功"})
}
