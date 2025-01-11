package config

import (
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"

	"CutMe/domain"
	"CutMe/infrastructure"
	"CutMe/usecase"
)

// Dependencies centraliza as instâncias necessárias na aplicação.
type Dependencies struct {
	S3Client           domain.S3Client
	DynamoClient       domain.DynamoClient
	EmailNotifier      domain.Notifier
	ProcessFileUseCase domain.ProcessFileUseCase
	SQSConsumer        domain.SQSConsumer
	SignedURLGenerator domain.SignedURLGenerator
}

func InitializeDependencies(awsSession *session.Session) *Dependencies {
	s3Bucket := getEnv("S3_BUCKET", "my-default-bucket")
	tableName := getEnv("DYNAMO_TABLE", "UploadsTable")
	queueURL := getEnv("QUEUE_URL", "https://sqs.sa-east-1.amazonaws.com/123456789012/minha-fila")
	cdnDomain := os.Getenv("CLOUDFRONT_DOMAIN_NAME") // pode ser vazio

	emailConfig := infrastructure.EmailConfig{
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     parseEnvInt(os.Getenv("SMTP_PORT"), 587),
		FromEmail:    getEnv("SMTP_EMAIL", "default@gmail.com"),
		FromPassword: os.Getenv("SMTP_PASSWORD"),
		ProjectName:  "CutMe",
	}

	// Criando implementações concretas
	s3Client := infrastructure.NewS3Client(awsSession)
	dynamoClient := infrastructure.NewDynamoClient(awsSession, tableName)
	emailNotifier := infrastructure.NewEmailNotifier(emailConfig)

	processFileUseCase := usecase.NewProcessFileUseCase(
		s3Client,
		s3Bucket,
		cdnDomain,
		dynamoClient,
		emailNotifier,
	)

	sqsClient := sqs.New(awsSession)

	// Nota: domain.SQSConsumer é a interface, e NewSQSConsumer retorna a implementação.
	sqsConsumer := infrastructure.NewSQSConsumer(
		sqsClient,
		s3Client,
		queueURL,
		processFileUseCase, // Esse "handler" implementa domain.SQSMessageHandler
	)

	signedURLGenerator := infrastructure.NewS3SignedURLGenerator(
		awsSession,
		s3Bucket,
		15*time.Minute,
	)

	return &Dependencies{
		S3Client:           s3Client,
		DynamoClient:       dynamoClient,
		EmailNotifier:      emailNotifier,
		ProcessFileUseCase: processFileUseCase,
		SQSConsumer:        sqsConsumer,
		SignedURLGenerator: signedURLGenerator,
	}
}

// parseEnvInt converte string -> int (fallback se der erro).
func parseEnvInt(value string, defaultValue int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// getEnv retorna variável de ambiente ou fallback se não existir.
func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
