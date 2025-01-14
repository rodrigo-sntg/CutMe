package signed_url

import (
	"CutMe/domain/repository"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

// S3SignedURLGenerator implementa a interface SignedURLGenerator
type S3SignedURLGenerator struct {
	s3Client    *s3.S3
	bucketName  string
	urlValidity time.Duration
}

// NewS3SignedURLGenerator cria uma nova instância do gerador de URLs assinadas
func NewS3SignedURLGenerator(sess *session.Session, bucketName string, urlValidity time.Duration) *S3SignedURLGenerator {
	return &S3SignedURLGenerator{
		s3Client:    s3.New(sess),
		bucketName:  bucketName,
		urlValidity: urlValidity,
	}
}

// GeneratePreSignedURL gera uma URL assinada para upload
func (g *S3SignedURLGenerator) GeneratePreSignedURL(fileName, fileType, userID string) (string, string, error) {
	uniqueID := uuid.New().String()
	uniqueFileName := fmt.Sprintf("%s_%s", uniqueID, fileName)

	req, _ := g.s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(g.bucketName),
		Key:         aws.String(uniqueFileName),
		ContentType: aws.String(fileType),
		Metadata: map[string]*string{
			"userID":   aws.String(userID),
			"uniqueID": aws.String(uniqueID),
		},
	})

	signedURL, err := req.Presign(g.urlValidity)
	if err != nil {
		return "", uniqueID, fmt.Errorf("erro ao gerar URL assinada: %w", err)
	}

	return signedURL, uniqueID, nil
}

// GetBucketName retorna o nome do bucket
func (g *S3SignedURLGenerator) GetBucketName() string {
	return g.bucketName
}

// GetURLValidity retorna a duração de validade da URL
func (g *S3SignedURLGenerator) GetURLValidity() time.Duration {
	return g.urlValidity
}

// GenerateUploadURLHandler é o handler que recebe o SignedURLGenerator como parâmetro de injeção
func GenerateUploadURLHandler(generator repository.SignedURLGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		type RequestBody struct {
			FileName string `json:"fileName" binding:"required"`
			FileType string `json:"fileType" binding:"required"`
		}

		var req RequestBody
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetros inválidos"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
			return
		}

		signedURL, uniqueID, err := generator.GeneratePreSignedURL(req.FileName, req.FileType, userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao gerar URL assinada"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"signedUrl": signedURL,
			"uniqueId":  uniqueID,
		})
	}
}

func UploadsHandler(dynamoClient repository.DynamoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists || userID == "" {
			c.JSON(401, gin.H{"error": "Usuário não autenticado"})
			return
		}

		status := c.Query("status")

		uploads, err := dynamoClient.GetUploadsByUserID(userID.(string), status)
		if err != nil {
			c.JSON(500, gin.H{"error": "Erro ao buscar uploads"})
			return
		}

		c.JSON(200, uploads)
	}
}

func CreateUploadHandler(dynamoClient repository.DynamoClient, s3Client repository.S3Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		type RequestBody struct {
			FileName string `json:"fileName" binding:"required"`
			FileType string `json:"fileType" binding:"required"`
		}

		var req RequestBody
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Parâmetros inválidos"})
			return
		}

		c.JSON(201, gin.H{"message": "Upload criado com sucesso"})
	}
}
