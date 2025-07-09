package router

import (
	"filesys/endpoint"
	"filesys/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	// 添加审计日志中间件
	r.Use(middleware.AuditLoggerMiddleware())

	// 静态文件服务（用于文件下载）
	r.Static("/data", "./data")

	// 登录路由
	r.POST("/login", endpoint.Login)

	// 需要认证的路由组
	auth := r.Group("/api", middleware.AuthMiddleware())
	{
		// 用户管理
		auth.POST("/user", endpoint.CreateUser)

		// 文件操作路由组（需要文件权限）
		fileGroup := auth.Group("/file", middleware.FilePermissionMiddleware())
		{
			// 文件夹操作
			fileGroup.POST("/:file_id/new", endpoint.CreateFolder)

			// 文件操作
			fileGroup.POST("/:file_id/upload", endpoint.UploadFile)
			fileGroup.POST("/:file_id/update", endpoint.UpdateFile)
			fileGroup.DELETE("/:file_id", endpoint.DeleteFile)
			fileGroup.POST("/:file_id/copy", endpoint.CopyFile)
			fileGroup.POST("/:file_id/move", endpoint.MoveFile)
			fileGroup.POST("/:file_id/rename", endpoint.RenameFile)

			// 文件查询
			fileGroup.GET("/:file_id", endpoint.GetFileInfo)
			fileGroup.GET("/:file_id/list", endpoint.ListFiles)
			fileGroup.GET("/:file_id/content", endpoint.DownloadFile)
			fileGroup.GET("/:file_id/version/:ver_num/content", endpoint.DownloadVersion)
		}
	}

	return r
}
