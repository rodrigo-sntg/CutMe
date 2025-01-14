package s3

import (
	"CutMe/internal/application/repository"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type s3Client struct {
	svc *s3.S3
}

func NewS3Client(sess *session.Session) repository.StorageClient {
	return &s3Client{
		svc: s3.New(sess),
	}
}

func (s *s3Client) DownloadFile(bucket, key string) (string, error) {
	decodedKey, err := url.QueryUnescape(key)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar chave do arquivo: %w", err)
	}

	out, err := s.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(decodedKey),
	})
	if err != nil {
		return "", fmt.Errorf("erro ao fazer download do S3: %w", err)
	}
	defer out.Body.Close()

	localFile := filepath.Join(os.TempDir(), key)
	f, err := os.Create(localFile)
	if err != nil {
		return "", fmt.Errorf("erro ao criar arquivo local: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, out.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao copiar bytes do S3: %w", err)
	}

	log.Printf("Arquivo baixado em: %s\n", localFile)
	return localFile, nil
}

func (s *s3Client) UploadFile(bucket, key, localPath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo para upload: %w", err)
	}
	defer f.Close()

	_, err = s.svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   f,
		Metadata: map[string]*string{
			"IgnoreSQS": aws.String("true"),
		},
	})
	if err != nil {
		return fmt.Errorf("erro no upload do arquivo: %w", err)
	}

	log.Printf("Upload concluído: s3://%s/%s\n", bucket, key)
	return nil
}

func (s *s3Client) HeadObject(bucket, key string) (*s3.HeadObjectOutput, error) {
	decodedKey, err := url.QueryUnescape(key)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar chave do arquivo: %w", err)
	}

	out, err := s.svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(decodedKey),
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao obter metadados do S3: %w", err)
	}
	return out, nil
}
