package config

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// SetupAWSSession cria a sessão AWS usando a região definida em AWS_REGION
func SetupAWSSession() *session.Session {
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}
	return session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion)}))
}
