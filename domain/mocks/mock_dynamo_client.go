package mocks

import (
	"CutMe/domain"

	"github.com/stretchr/testify/mock"
)

type MockDynamoClient struct {
	mock.Mock
}

func (m *MockDynamoClient) UpdateStatus(id, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockDynamoClient) UpdateUploadRecord(record domain.UploadRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *MockDynamoClient) CreateUploadRecord(record domain.UploadRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *MockDynamoClient) CreateOrUpdateUploadRecord(record domain.UploadRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *MockDynamoClient) GetUploads(status string) ([]domain.UploadRecord, error) {
	args := m.Called(status)
	return args.Get(0).([]domain.UploadRecord), args.Error(1)
}

func (m *MockDynamoClient) GetUploadByID(id string) (*domain.UploadRecord, error) {
	args := m.Called(id)
	// Caso deseje retornar um ponteiro nulo, faça um cast seguro com verificação
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*domain.UploadRecord), args.Error(1)
}

func (m *MockDynamoClient) GetUploadsByUserID(userID string, status string) ([]domain.UploadRecord, error) {
	args := m.Called(userID, status)
	return args.Get(0).([]domain.UploadRecord), args.Error(1)
}
