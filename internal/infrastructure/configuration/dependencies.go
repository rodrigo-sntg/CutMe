package configuration

import (
	repository2 "CutMe/internal/application/repository"
	service2 "CutMe/internal/application/service"
	usecase2 "CutMe/internal/application/usecase"
	"CutMe/internal/infrastructure/aws/db"
	"CutMe/internal/infrastructure/aws/email"
	"CutMe/internal/infrastructure/aws/s3"
	"CutMe/internal/infrastructure/aws/signed_url"
	sqs2 "CutMe/internal/infrastructure/aws/sqs"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// Dependencies centraliza as instâncias necessárias na aplicação.
type Dependencies struct {
	S3Client           repository2.StorageClient
	DynamoClient       repository2.DBClient
	EmailNotifier      service2.Notifier
	ProcessFileUseCase usecase2.ProcessFileUseCase
	SQSConsumer        service2.QueueConsumer
	SignedURLGenerator repository2.PresignedURLGenerator
}

func InitializeDependencies(awsSession *session.Session) *Dependencies {
	s3Bucket := getEnv("S3_BUCKET", "my-default-bucket")
	tableName := getEnv("DYNAMO_TABLE", "UploadsTable")
	queueURL := getEnv("QUEUE_URL", "https://sqs.sa-east-1.amazonaws.com/123456789012/minha-fila")
	cdnDomain := os.Getenv("CLOUDFRONT_DOMAIN_NAME") // pode ser vazio

	emailConfig := email.EmailConfig{
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     parseEnvInt(os.Getenv("SMTP_PORT"), 587),
		FromEmail:    getEnv("SMTP_EMAIL", "default@gmail.com"),
		FromPassword: os.Getenv("SMTP_PASSWORD"),
		ProjectName:  "CutMe",
	}

	// Criando implementações concretas
	s3Client := s3.NewS3Client(awsSession)
	dynamoClient := db.NewDynamoClient(awsSession, tableName)
	emailNotifier := email.NewEmailNotifier(emailConfig)

	extractFrames := func(localVideo string) (string, error) {
		return usecase2.DefaultExtractFrames(localVideo)
	}

	zipFrames := func(framesDir string) (string, error) {
		return usecase2.DefaultZipFrames(framesDir)
	}

	processFileUseCase := usecase2.NewProcessFileUseCase(
		s3Client,
		s3Bucket,
		cdnDomain,
		dynamoClient,
		emailNotifier,
		extractFrames,
		zipFrames,
	)

	sqsClient := sqs.New(awsSession)

	// Nota: domain.QueueConsumer é a interface, e NewSQSConsumer retorna a implementação.
	sqsConsumer := sqs2.NewSQSConsumer(
		sqsClient,
		s3Client,
		queueURL,
		processFileUseCase, // Esse "handler" implementa domain.SQSMessageHandler
	)

	signedURLGenerator := signed_url.NewS3SignedURLGenerator(
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
