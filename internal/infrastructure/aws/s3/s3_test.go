package s3_test

import (
	"CutMe/internal/application/mocks"
	cutmeS3 "CutMe/internal/infrastructure/aws/s3" // alias para evitar colisão no nome do pacote "s3"
	"bytes"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/mock/gomock"
)

// Se ainda não tiver, gere o mock com:
// mockgen -destination=internal/application/mocks/mock_s3.go -package=mocks github.com/aws/aws-sdk-go/service/s3/s3iface S3API

func TestDownloadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockS3 := mocks.NewMockS3API(ctrl)
	client := &cutmeS3.S3Client{Svc: mockS3}

	t.Run("success", func(t *testing.T) {
		bucket := "test-bucket"
		key := "test-key"
		decodedKey := key
		tempFile := filepath.Join(os.TempDir(), key)

		mockS3.EXPECT().
			GetObject(gomock.Any()).
			DoAndReturn(func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
				// Verifica se os parâmetros são os esperados
				if *input.Bucket != bucket || *input.Key != decodedKey {
					t.Errorf("unexpected input: %v", input)
				}
				// Retorna um objeto com "file content" no Body
				return &s3.GetObjectOutput{
					Body: ioutil.NopCloser(bytes.NewReader([]byte("file content"))),
				}, nil
			})

		file, err := client.DownloadFile(bucket, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if file != tempFile {
			t.Fatalf("expected %s, got %s", tempFile, file)
		}
	})

	t.Run("error decoding key", func(t *testing.T) {
		invalidKey := "%invalid"

		_, err := client.DownloadFile("bucket", invalidKey)
		if err == nil || !strings.Contains(err.Error(), "invalid URL escape") {
			t.Fatalf("expected decoding error, got %v", err)
		}
	})

	t.Run("error downloading file", func(t *testing.T) {
		mockS3.EXPECT().
			GetObject(gomock.Any()).
			Return(nil, errors.New("download error"))

		_, err := client.DownloadFile("bucket", "key")
		if err == nil || err.Error() != "erro ao fazer download do S3: download error" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestUploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockS3 := mocks.NewMockS3API(ctrl)
	client := &cutmeS3.S3Client{Svc: mockS3}

	t.Run("success", func(t *testing.T) {
		bucket := "test-bucket"
		key := "test-key"
		localPath := "test-file.txt"
		fileContent := []byte("test content")

		// Cria um arquivo temporário para testar
		err := ioutil.WriteFile(localPath, fileContent, 0644)
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(localPath)

		mockS3.EXPECT().
			PutObject(gomock.Any()).
			DoAndReturn(func(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
				if *input.Bucket != bucket || *input.Key != key {
					t.Errorf("unexpected input: %v", input)
				}
				return &s3.PutObjectOutput{}, nil
			})

		err = client.UploadFile(bucket, key, localPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error opening file", func(t *testing.T) {
		err := client.UploadFile("bucket", "key", "non-existent-file.txt")
		if err == nil || !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("expected file not exist error, got %v", err)
		}
	})

	t.Run("error uploading file", func(t *testing.T) {
		// Cria um arquivo temporário para o teste
		localPath := "temp-test-file.txt"
		err := ioutil.WriteFile(localPath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(localPath) // Remove o arquivo após o teste

		mockS3.EXPECT().
			PutObject(gomock.Any()).
			Return(nil, errors.New("upload error"))

		err = client.UploadFile("bucket", "key", localPath)
		if err == nil || err.Error() != "erro no upload do arquivo: upload error" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestHeadObject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockS3 := mocks.NewMockS3API(ctrl)
	client := &cutmeS3.S3Client{Svc: mockS3}

	t.Run("success", func(t *testing.T) {
		bucket := "test-bucket"
		key := "test-key"

		mockS3.EXPECT().
			HeadObject(gomock.Any()).
			DoAndReturn(func(input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
				if *input.Bucket != bucket || *input.Key != key {
					t.Errorf("unexpected input: %v", input)
				}
				return &s3.HeadObjectOutput{}, nil
			})

		_, err := client.HeadObject(bucket, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error decoding key", func(t *testing.T) {
		invalidKey := "%invalid"

		_, err := client.DownloadFile("bucket", invalidKey)
		if err == nil || !strings.Contains(err.Error(), "invalid URL escape") {
			t.Fatalf("expected decoding error, got %v", err)
		}
	})

	t.Run("error retrieving metadata", func(t *testing.T) {
		mockS3.EXPECT().
			HeadObject(gomock.Any()).
			Return(nil, errors.New("metadata error"))

		_, err := client.HeadObject("bucket", "key")
		if err == nil || err.Error() != "erro ao obter metadados do S3: metadata error" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
func TestNewS3Client(t *testing.T) {
	t.Run("successfully creates S3 client", func(t *testing.T) {
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String("us-east-1"),
		})
		if err != nil {
			t.Fatalf("failed to create AWS session: %v", err)
		}

		client := cutmeS3.NewS3Client(sess)
		if client == nil {
			t.Fatal("expected non-nil S3Client")
		}

		typedClient, ok := client.(*cutmeS3.S3Client)
		if !ok {
			t.Fatalf("expected type S3Client, got %T", client)
		}

		if typedClient.Svc == nil {
			t.Fatal("expected non-nil Svc in S3Client")
		}
	})
}
