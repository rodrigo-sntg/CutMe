package config

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func SetupAWSSession() *session.Session {
	awsRegion := os.Getenv("AWS_REGION")
	return session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion)}))
}
