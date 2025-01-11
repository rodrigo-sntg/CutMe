package routes

import (
	"github.com/gin-gonic/gin"

	"CutMe/domain"
	"CutMe/infrastructure"
	"CutMe/middleware"
)

type Dependencies struct {
	DynamoClient       domain.DynamoClient
	S3Client           domain.S3Client
	SignedURLGenerator domain.SignedURLGenerator
}

// RegisterRoutes registra as rotas da aplicação.
func RegisterRoutes(router *gin.Engine, deps *Dependencies) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Bem-vindo à API CutMe"})
	})

	// Rotas autenticadas
	authGroup := router.Group("/api")
	authGroup.Use(middleware.AuthMiddleware())
	{
		// Invocamos os handlers do pacote "infrastructure",
		// mas poderíamos criar outro pacote "handlers" se quisermos.
		authGroup.GET("/uploads", infrastructure.UploadsHandler(deps.DynamoClient))
		authGroup.POST("/upload", infrastructure.CreateUploadHandler(deps.DynamoClient, deps.S3Client))
		authGroup.POST("/uploads/signed-url", infrastructure.GenerateUploadURLHandler(deps.SignedURLGenerator))
	}
}
