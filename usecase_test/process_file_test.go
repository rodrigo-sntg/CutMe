package usecase_test

import (
	"CutMe/domain"
	"CutMe/usecase"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

type MockDynamoClient struct {
	mock.Mock
}

func (m *MockDynamoClient) UpdateStatus(id, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) SendSuccessEmail(to, uploadID string) error {
	args := m.Called(to, uploadID)
	return args.Error(0)
}

func (m *MockNotifier) SendFailureEmail(to, uploadID, errorMsg string) error {
	args := m.Called(to, uploadID, errorMsg)
	return args.Error(0)
}

func TestProcessFileUseCase_Handle_Success(t *testing.T) {
	mockS3 := &MockS3Client{}
	mockDynamo := &MockDynamoClient{}
	mockNotifier := &MockNotifier{}

	// Configurar mocks
	mockS3.On("DownloadFile", "test-bucket", "test-file").Return("/tmp/test-file", nil)
	mockS3.On("UploadFile", "test-bucket", "test-file_processed", "/tmp/test-file").Return(nil)
	mockDynamo.On("UpdateStatus", "123", "PROCESSING").Return(nil)
	mockDynamo.On("UpdateStatus", "123", "PROCESSED").Return(nil)
	mockNotifier.On("SendSuccessEmail", "user@example.com", "123").Return(nil)

	// Criar caso de uso
	uc := usecase.NewProcessFileUseCase(mockS3, "test-bucket", mockDynamo, mockNotifier)

	// Mensagem SQS simulada
	msg := domain.SQSMessage{
		ID:        "123",
		FileName:  "test-file",
		Bucket:    "test-bucket",
		UserEmail: "user@example.com",
	}

	err := uc.Handle(context.Background(), msg)
	require.NoError(t, err)

	// Validar chamadas
	mockS3.AssertExpectations(t)
	mockDynamo.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}

func TestProcessFileUseCase_Handle_Failure(t *testing.T) {
	mockS3 := &MockS3Client{}
	mockDynamo := &MockDynamoClient{}
	mockNotifier := &MockNotifier{}

	// Configurar mocks
	mockS3.On("DownloadFile", "test-bucket", "test-file").Return("", errors.New("download error"))
	mockDynamo.On("UpdateStatus", "123", "PROCESSING").Return(nil)
	mockNotifier.On("SendFailureEmail", "user@example.com", "123", "Erro ao baixar o arquivo: download error").Return(nil)

	// Criar caso de uso
	uc := usecase.NewProcessFileUseCase(mockS3, "test-bucket", mockDynamo, mockNotifier)

	// Mensagem SQS simulada
	msg := domain.SQSMessage{
		ID:        "123",
		FileName:  "test-file",
		Bucket:    "test-bucket",
		UserEmail: "user@example.com",
	}

	err := uc.Handle(context.Background(), msg)
	require.Error(t, err)

	// Validar chamadas
	mockS3.AssertExpectations(t)
	mockDynamo.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}
