package http

import (
	"CutMe/domain/repository"
	"CutMe/infrastructure/signed_url"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	DynamoClient       repository.DynamoClient
	S3Client           repository.S3Client
	SignedURLGenerator repository.SignedURLGenerator
}

// RegisterRoutes registra as rotas da aplicação.
func RegisterRoutes(router *gin.Engine, deps *Dependencies) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Bem-vindo à API CutMe"})
	})

	// Rotas autenticadas
	authGroup := router.Group("/api")
	authGroup.Use(AuthMiddleware())
	{
		// Invocamos os handlers do pacote "infrastructure",
		// mas poderíamos criar outro pacote "handlers" se quisermos.
		authGroup.GET("/uploads", signed_url.UploadsHandler(deps.DynamoClient))
		authGroup.POST("/upload", signed_url.CreateUploadHandler(deps.DynamoClient, deps.S3Client))
		authGroup.POST("/uploads/signed-url", signed_url.GenerateUploadURLHandler(deps.SignedURLGenerator))
	}
}
