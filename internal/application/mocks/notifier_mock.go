package mocks

import (
	"github.com/stretchr/testify/mock"
)

type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) SendSuccessEmailWithLinks(userID, uploadID, originalFileLink, processedFileLink string) error {
	args := m.Called(userID, uploadID, originalFileLink, processedFileLink)
	return args.Error(0)
}

func (m *NotifierMock) SendFailureEmail(userID, uploadID, reason string) error {
	args := m.Called(userID, uploadID, reason)
	return args.Error(0)
}
