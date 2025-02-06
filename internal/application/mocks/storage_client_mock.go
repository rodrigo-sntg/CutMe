// storage_client_mock.go
package mocks

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/mock"
)

type StorageClientMock struct {
	mock.Mock
}

func (m *StorageClientMock) DownloadFile(bucket, key string) (string, error) {
	args := m.Called(bucket, key)
	return args.String(0), args.Error(1)
}

func (m *StorageClientMock) UploadFile(bucket, key, filePath string) error {
	args := m.Called(bucket, key, filePath)
	return args.Error(0)
}

func (m *StorageClientMock) HeadObject(bucket, key string) (*s3.HeadObjectOutput, error) {
	args := m.Called(bucket, key)
	// se quiser simular retorno *s3.HeadObjectOutput, coloque algo como:
	// return args.Get(0).(*s3.HeadObjectOutput), args.Error(1)
	// mas se não usar esse método no seu teste, pode só retornar nil, nil
	// ou definir de outra forma. Exemplo:
	return nil, args.Error(1)
}
