package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"filesys/dao"
	"filesys/model"
	"filesys/service"
	"filesys/storage"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Login 登录接口
func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// 计算 SHA-256 哈希
	hash := sha256.Sum256([]byte(password))
	hashStr := hex.EncodeToString(hash[:])

	user, err := dao.Q.User.Where(dao.Q.User.Name.Eq(username), dao.Q.User.Password.Eq(hashStr)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(401, gin.H{"error": "用户名或密码错误"})
		} else {
			c.JSON(500, gin.H{"error": "数据库错误"})
		}
		return
	}

	// 生成 sid
	sid := generateSID()

	// 保存会话信息
	// 这里可以实现会话存储逻辑，例如使用数据库或缓存

	c.SetCookie("sid", sid, 3600, "/", "", false, true)
	c.JSON(200, gin.H{"sid": sid})
}

// CreateUser 创建用户
func CreateUser(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// 计算 SHA-256 哈希
	hash := sha256.Sum256([]byte(password))
	hashStr := hex.EncodeToString(hash[:])

	err := service.CreateUserService.CreateUser(context.Background(), username, hashStr)
	if err != nil {
		c.JSON(500, gin.H{"error": "创建用户失败"})
		return
	}

	c.JSON(200, gin.H{"message": "用户创建成功"})
}

// CreateFolder 新建文件夹
func CreateFolder(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的文件ID"})
		return
	}

	// 获取当前用户ID
	// 这里可以从会话信息中获取用户ID
	userID := int32(1) // 暂时写死

	// 实现新建文件夹逻辑
	folderName := c.PostForm("name")
	// 处理重名问题
	folderName = handleDuplicateName(fileID, folderName, "folder")

	folder := &model.File{
		UserID:   userID,
		ParentID: int32(fileID),
		Name:     folderName,
		Size:     0,
		Type:     "folder",
		VerNum:   0,
		StoreKey: "",
		Ctime:    int32(time.Now().Unix()),
		Mtime:    int32(time.Now().Unix()),
	}

	err = dao.Q.File.Create(folder)
	if err != nil {
		c.JSON(500, gin.H{"error": "创建文件夹失败"})
		return
	}

	c.JSON(200, folder)
}

// UploadFile 上传文件
func UploadFile(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的文件ID"})
		return
	}

	// 获取当前用户ID
	// 这里可以从会话信息中获取用户ID
	userID := int32(1) // 暂时写死

	// 读取文件内容
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "未找到文件"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "打开文件失败"})
		return
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		c.JSON(500, gin.H{"error": "读取文件内容失败"})
		return
	}

	// 处理重名问题
	fileName := file.Filename
	fileName = handleDuplicateName(fileID, fileName, "file")

	// 保存文件到磁盘
	storeKey, err := storage.SaveFile(content)
	if err != nil {
		c.JSON(500, gin.H{"error": "保存文件到磁盘失败"})
		return
	}

	// 创建文件记录
	newFile := &model.File{
		UserID:   userID,
		ParentID: int32(fileID),
		Name:     fileName,
		Size:     int32(len(content)),
		Type:     "file",
		VerNum:   1,
		StoreKey: storeKey,
		Ctime:    int32(time.Now().Unix()),
		Mtime:    int32(time.Now().Unix()),
	}

	err = dao.Q.File.Create(newFile)
	if err != nil {
		c.JSON(500, gin.H{"error": "创建文件记录失败"})
		return
	}

	// 创建版本记录
	newVersion := &model.Version{
		FileID:   newFile.ID,
		Size:     int32(len(content)),
		VerNum:   1,
		StoreKey: storeKey,
		Ctime:    int32(time.Now().Unix()),
	}

	err = dao.Q.Version.Create(newVersion)
	if err != nil {
		c.JSON(500, gin.H{"error": "创建版本记录失败"})
		return
	}

	c.JSON(200, newFile)
}

// UpdateFile 更新文件
func UpdateFile(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的文件ID"})
		return
	}

	// 读取文件内容
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "未找到文件"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "打开文件失败"})
		return
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		c.JSON(500, gin.H{"error": "读取文件内容失败"})
		return
	}

	// 获取原文件信息
	srcFile, err := dao.Q.File.Where(dao.Q.File.ID.Eq(int32(fileID))).First()
	if err != nil {
		c.JSON(404, gin.H{"error": "文件不存在"})
		return
	}

	// 保存文件到磁盘
	storeKey, err := storage.SaveFile(content)
	if err != nil {
		c.JSON(500, gin.H{"error": "保存文件到磁盘失败"})
		return
	}

	// 更新文件记录
	newVerNum := srcFile.VerNum + 1
	_, err = dao.Q.File.Where(dao.Q.File.ID.Eq(int32(fileID))).Update(
		dao.Q.File.Size, int32(len(content)),
		dao.Q.File.VerNum, newVerNum,
		dao.Q.File.StoreKey, storeKey,
		dao.Q.File.Mtime, int32(time.Now().Unix()),
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "更新文件记录失败"})
		return
	}

	// 创建新版本记录
	newVersion := &model.Version{
		FileID:   srcFile.ID,
		Size:     int32(len(content)),
		VerNum:   newVerNum,
		StoreKey: storeKey,
		Ctime:    int32(time.Now().Unix()),
	}

	err = dao.Q.Version.Create(newVersion)
	if err != nil {
		c.JSON(500, gin.H{"error": "创建版本记录失败"})
		return
	}

	updatedFile, err := dao.Q.File.Where(dao.Q.File.ID.Eq(int32(fileID))).First()
	if err != nil {
		c.JSON(500, gin.H{"error": "获取更新后的文件信息失败"})
		return
	}

	c.JSON(200, updatedFile)
}

// DeleteFile 删除文件（夹）
func DeleteFile(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的文件ID"})
		return
	}

	// 获取文件信息
	file, err := dao.Q.File.Where(dao.Q.File.ID.Eq(int32(fileID))).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"error": "文件不存在"})
		} else {
			c.JSON(500, gin.H{"error": "数据库错误"})
		}
		return
	}

	// 开始事务
	tx := dao.Q.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除文件（夹）
	if file.Type == "folder" {
		// 递归删除子文件（夹）
		err := deleteFolderContents(tx, int32(fileID))
		if err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "删除文件夹内容失败"})
			return
		}
	} else {
		// 减少存储引用
		if err := updateStoreRef(tx, file.StoreKey, -1); err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "减少存储引用失败"})
			return
		}

		// 删除版本记录
		_, err := tx.Version.Where(tx.Version.FileID.Eq(file.ID)).Delete()
		if err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "删除版本记录失败"})
			return
		}
	}

	// 删除文件记录
	_, err = tx.File.Where(tx.File.ID.Eq(file.ID)).Delete()
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "删除文件记录失败"})
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(500, gin.H{"error": "提交事务失败"})
		return
	}

	c.JSON(200, gin.H{"message": "文件（夹）删除成功"})
}

// CopyFile 复制文件（夹）
func CopyFile(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的文件ID"})
		return
	}

	newParentIDStr := c.PostForm("new_parent_id")
	newParentID, err := strconv.ParseInt(newParentIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的新父目录ID"})
		return
	}

	newFile, err := service.CopyService.CopyFile(context.Background(), uint(fileID), uint(newParentID))
	if err != nil {
		c.JSON(500, gin.H{"error": "复制文件（夹）失败: " + err.Error()})
		return
	}

	c.JSON(200, newFile)
}

// MoveFile 移动文件（夹）
func MoveFile(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的文件ID"})
		return
	}

	newParentIDStr := c.PostForm("new_parent_id")
	newParentID, err := strconv.ParseInt(newParentIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的新父目录ID"})
		return
	}

	// 开始事务
	tx := dao.Q.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新文件（夹）的父目录
	_, err = tx.File.Where(tx.File.ID.Eq(int32(fileID))).Update(tx.File.ParentID, int32(newParentID))
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "更新文件（夹）父目录失败"})
		return
	}

	// 如果是文件夹，递归更新子文件路径
	file, err := tx.File.Where(tx.File.ID.Eq(int32(fileID))).First()
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "获取文件（夹）信息失败"})
		return
	}

	if file.Type == "folder" {
		err := service.UpdateChildPaths(tx, uint(file.ID), uint(newParentID))
		if err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "递归更新子文件路径失败"})
			return
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(500, gin.H{"error": "提交事务失败"})
		return
	}

	movedFile, err := dao.Q.File.Where(dao.Q.File.ID.Eq(int32(fileID))).First()
	if err != nil {
		c.JSON(500, gin.H{"error": "获取移动后的文件（夹）信息失败"})
		return
	}

	c.JSON(200, movedFile)
}

// RenameFile 重命名文件（夹）
func RenameFile(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的文件ID"})
		return
	}

	newName := c.PostForm("new_name")
	if newName == "" {
		c.JSON(400, gin.H{"error": "新名称不能为空"})
		return
	}

	// 获取文件信息
	file, err := dao.Q.File.Where(dao.Q.File.ID.Eq(int32(fileID))).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"error": "文件不存在"})
		} else {
			c.JSON(500, gin.H{"error": "数据库错误"})
		}
		return
	}

	// 处理重名问题
	newName = handleDuplicateName(file.ParentID, newName, file.Type)

	// 更新文件（夹）名称
	_, err = dao.Q.File.Where(dao.Q.File.ID.Eq(int32(fileID))).Update(dao.Q.File.Name, newName)
	if err != nil {
		c.JSON(500, gin.H{"error": "重命名文件（夹）失败"})
		return
	}

	renamedFile, err := dao.Q.File.Where(dao.Q.File.ID.Eq(int32(fileID))).First()
	if err != nil {
		c.JSON(500, gin.H{"error": "获取重命名后的文件（夹）信息失败"})
		return
	}

	c.JSON(200, renamedFile)
}

// GetFile 获取文件（夹）信息
func GetFile(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的文件ID"})
		return
	}

	file, err := dao.Q.File.Where(dao.Q.File.ID.Eq(int32(fileID))).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"error": "文件不存在"})
		} else {
			c.JSON(500, gin.H{"error": "数据库错误"})
		}
		return
	}

	c.JSON(200, file)
}

// ListFiles 列出子文件（夹）
func ListFiles(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的文件ID"})
		return
	}

	files, err := dao.Q.File.Where(dao.Q.File.ParentID.Eq(int32(fileID))).Find()
	if err != nil {
		c.JSON(500, gin.H{"error": "获取子文件（夹）列表失败"})
		return
	}

	c.JSON(200, files)
}

// DownloadFile 下载文件
func DownloadFile(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的文件ID"})
		return
	}

	// 获取文件信息
	file, err := dao.Q.File.Where(dao.Q.File.ID.Eq(int32(fileID))).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"error": "文件不存在"})
		} else {
			c.JSON(500, gin.H{"error": "数据库错误"})
		}
		return
	}

	if file.Type != "file" {
		c.JSON(400, gin.H{"error": "该ID对应的不是文件"})
		return
	}

	// 构建文件路径
	path := storage.GetFilePath(file.StoreKey)

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": "文件不存在于磁盘"})
		return
	}

	// 下载文件
	c.File(path)
}

// DownloadHistoricalVersion 下载历史版本
func DownloadHistoricalVersion(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的文件ID"})
		return
	}

	verNumStr := c.Param("ver_num")
	verNum, err := strconv.ParseInt(verNumStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的版本号"})
		return
	}

	// 获取版本信息
	version, err := dao.Q.Version.Where(dao.Q.Version.FileID.Eq(int32(fileID)), dao.Q.Version.VerNum.Eq(int32(verNum))).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"error": "版本记录不存在"})
		} else {
			c.JSON(500, gin.H{"error": "数据库错误"})
		}
		return
	}

	// 构建文件路径
	path := storage.GetFilePath(version.StoreKey)

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": "文件不存在于磁盘"})
		return
	}

	// 下载文件
	c.File(path)
}

// 处理重名问题
func handleDuplicateName(parentID int32, name string, fileType string) string {
	count := 1
	originalName := name
	for {
		existingFiles, err := dao.Q.File.Where(dao.Q.File.ParentID.Eq(parentID), dao.Q.File.Name.Eq(name)).Find()
		if err != nil {
			break
		}
		if len(existingFiles) == 0 {
			break
		}
		if strings.Contains(originalName, " (") {
			originalName = strings.Split(originalName, " (")[0]
		}
		name = fmt.Sprintf("%s (%d)", originalName, count)
		count++
	}
	return name
}

// 生成 sid
func generateSID() string {
	// 这里可以实现更复杂的 sid 生成逻辑
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

// 递归删除文件夹内容
func deleteFolderContents(tx *dao.Query, folderID int32) error {
	// 获取所有子文件
	children, err := tx.File.Where(tx.File.ParentID.Eq(folderID)).Find()
	if err != nil {
		return err
	}

	// 删除每个子文件
	for _, child := range children {
		if child.Type == "folder" {
			// 递归删除子文件夹
			if err := deleteFolderContents(tx, child.ID); err != nil {
				return err
			}
		} else {
			// 减少存储引用
			if err := updateStoreRef(tx, child.StoreKey, -1); err != nil {
				return err
			}

			// 删除版本记录
			_, err := tx.Version.Where(tx.Version.FileID.Eq(child.ID)).Delete()
			if err != nil {
				return err
			}
		}

		// 删除文件记录
		_, err := tx.File.Where(tx.File.ID.Eq(child.ID)).Delete()
		if err != nil {
			return err
		}
	}

	return nil
}

// 更新存储引用
func updateStoreRef(tx *dao.Query, storeKey string, delta int) error {
	storeRef, err := tx.StoreRef.Where(tx.StoreRef.StoreKey.Eq(storeKey)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if delta > 0 {
				// 创建新的存储引用记录
				newStoreRef := &model.StoreRef{
					StoreKey: storeKey,
					RefCount: int32(delta),
				}
				return tx.StoreRef.Create(newStoreRef)
			}
			return nil
		}
		return err
	}

	newRefCount := storeRef.RefCount + int32(delta)
	if newRefCount <= 0 {
		// 删除存储引用记录
		_, err := tx.StoreRef.Where(tx.StoreRef.StoreKey.Eq(storeKey)).Delete()
		if err != nil {
			return err
		}

		// 删除磁盘上的文件
		path := storage.GetFilePath(storeKey)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	} else {
		// 更新存储引用记录
		_, err := tx.StoreRef.Where(tx.StoreRef.StoreKey.Eq(storeKey)).Update(tx.StoreRef.RefCount, newRefCount)
		if err != nil {
			return err
		}
	}

	return nil
}
