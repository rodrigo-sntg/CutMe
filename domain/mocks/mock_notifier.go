package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) SendSuccessEmailWithLinks(to, uploadID, originalFileURL, processedFileURL string) error {
	args := m.Called(to, uploadID, originalFileURL, processedFileURL)
	return args.Error(0)
}

func (m *MockNotifier) SendFailureEmail(to, uploadID, errorMsg string) error {
	args := m.Called(to, uploadID, errorMsg)
	return args.Error(0)
}
