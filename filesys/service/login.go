package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"filesys/dao"
)

type loginService struct{}

var LoginService = &loginService{}

func (s *loginService) Login(ctx context.Context, username, password string) (int32, error) {
	user, err := dao.Q.User.WithContext(ctx).Where(dao.User.Name.Eq(username)).First()
	if err != nil || user == nil {
		return 0, errors.New("用户不存在")
	}

	// 计算 SHA-256 哈希
	hash := sha256.Sum256([]byte(password))

	// 转为十六进制字符串
	hashStr := hex.EncodeToString(hash[:])
	if hashStr != user.Password {
		return 0, errors.New("密码错误")
	}

	return user.ID, nil
}
