package routes

import (
	"CutMe/domain"
	"CutMe/infrastructure"
	"CutMe/middleware"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	DynamoClient       domain.DynamoClient
	S3Client           domain.S3Client
	SignedURLGenerator domain.SignedURLGenerator
}

func RegisterRoutes(router *gin.Engine, deps *Dependencies) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Bem-vindo à API CutMe"})
	})

	authGroup := router.Group("/api")
	authGroup.Use(middleware.AuthMiddleware())
	{
		authGroup.GET("/uploads", infrastructure.UploadsHandler(deps.DynamoClient))
		authGroup.POST("/upload", infrastructure.CreateUploadHandler(deps.DynamoClient, deps.S3Client))
		authGroup.POST("/uploads/signed-url", infrastructure.GenerateUploadURLHandler(deps.SignedURLGenerator))
	}
}
