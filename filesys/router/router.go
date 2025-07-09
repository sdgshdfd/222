package router

import (
	"filesys/controller"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	// 登录接口
	r.POST("/login", controller.Login)

	// 管理接口
	adminGroup := r.Group("/api/user")
	adminGroup.Use(controller.AdminAuthMiddleware())
	adminGroup.POST("", controller.CreateUser)

	// 文件接口
	fileGroup := r.Group("/api/file")
	fileGroup.Use(controller.SessionAuthMiddleware())
	fileGroup.POST("/:file_id/new", controller.CreateFolder)
	fileGroup.POST("/:file_id/upload", controller.UploadFile)
	fileGroup.POST("/:file_id/update", controller.UpdateFile)
	fileGroup.DELETE("/:file_id", controller.DeleteFile)
	fileGroup.POST("/:file_id/copy", controller.CopyFile)
	fileGroup.POST("/:file_id/move", controller.MoveFile)
	fileGroup.POST("/:file_id/rename", controller.RenameFile)
	fileGroup.GET("/:file_id", controller.GetFile)
	fileGroup.GET("/:file_id/list", controller.ListFiles)
	fileGroup.GET("/:file_id/content", controller.DownloadFile)
	fileGroup.GET("/:file_id/version/:ver_num/content", controller.DownloadHistoricalVersion)

	return r
}
