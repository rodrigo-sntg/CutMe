package mocks

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/mock"
)

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) DownloadFile(bucket, key string) (string, error) {
	args := m.Called(bucket, key)
	return args.String(0), args.Error(1)
}

func (m *MockS3Client) UploadFile(bucket, key, localPath string) error {
	args := m.Called(bucket, key, localPath)
	return args.Error(0)
}

func (m *MockS3Client) HeadObject(bucket, key string) (*s3.HeadObjectOutput, error) {
	args := m.Called(bucket, key)
	// Precisamos converter o resultado para *s3.HeadObjectOutput
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*s3.HeadObjectOutput), args.Error(1)
}
