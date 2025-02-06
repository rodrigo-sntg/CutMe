package mocks

import (
	"CutMe/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

type DBClientMock struct {
	mock.Mock
}

func (m *DBClientMock) UpdateStatus(id, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *DBClientMock) UpdateUploadRecord(record entity.UploadRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *DBClientMock) CreateUploadRecord(record entity.UploadRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *DBClientMock) CreateOrUpdateUploadRecord(record entity.UploadRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *DBClientMock) GetUploads(status string) ([]entity.UploadRecord, error) {
	args := m.Called(status)
	return args.Get(0).([]entity.UploadRecord), args.Error(1)
}

func (m *DBClientMock) GetUploadByID(id string) (*entity.UploadRecord, error) {
	args := m.Called(id)
	return args.Get(0).(*entity.UploadRecord), args.Error(1)
}

func (m *DBClientMock) GetUploadsByUserID(userID string, status string) ([]entity.UploadRecord, error) {
	args := m.Called(userID, status)
	return args.Get(0).([]entity.UploadRecord), args.Error(1)
}
