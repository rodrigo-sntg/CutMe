package config

import (
	"CutMe/domain"
	"CutMe/infrastructure"
	"CutMe/usecase"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Dependencies struct {
	S3Client           domain.S3Client
	DynamoClient       domain.DynamoClient
	EmailNotifier      domain.Notifier
	ProcessFileUseCase domain.ProcessFileUseCase
	SQSConsumer        *infrastructure.SQSConsumer
	SignedURLGenerator domain.SignedURLGenerator
}

func InitializeDependencies(awsSession *session.Session) *Dependencies {
	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		s3Bucket = "my-default-bucket"
	}

	tableName := os.Getenv("DYNAMO_TABLE")
	if tableName == "" {
		tableName = "UploadsTable"
	}

	queueURL := os.Getenv("QUEUE_URL")
	if queueURL == "" {
		queueURL = "https://sqs.sa-east-1.amazonaws.com/123456789012/minha-fila"
	}

	cdnDomain := os.Getenv("CLOUDFRONT_DOMAIN_NAME")

	emailConfig := infrastructure.EmailConfig{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     parseEnvInt(os.Getenv("SMTP_PORT"), 587),
		FromEmail:    os.Getenv("SMTP_EMAIL"),
		FromPassword: os.Getenv("SMTP_PASSWORD"),
		ProjectName:  "CutMe",
	}

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
	sqsConsumer := infrastructure.NewSQSConsumer(
		sqsClient,
		s3Client,
		queueURL,
		processFileUseCase,
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

func parseEnvInt(value string, defaultValue int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
