// Code generated by MockGen. DO NOT EDIT.
// Source: CutMe/internal/application/repository (interfaces: QueueClient)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	sqs "github.com/aws/aws-sdk-go/service/sqs"
	gomock "github.com/golang/mock/gomock"
)

// MockQueueClient is a mock of QueueClient interface.
type MockQueueClient struct {
	ctrl     *gomock.Controller
	recorder *MockQueueClientMockRecorder
}

// MockQueueClientMockRecorder is the mock recorder for MockQueueClient.
type MockQueueClientMockRecorder struct {
	mock *MockQueueClient
}

// NewMockQueueClient creates a new mock instance.
func NewMockQueueClient(ctrl *gomock.Controller) *MockQueueClient {
	mock := &MockQueueClient{ctrl: ctrl}
	mock.recorder = &MockQueueClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQueueClient) EXPECT() *MockQueueClientMockRecorder {
	return m.recorder
}

// DeleteMessage mocks base method.
func (m *MockQueueClient) DeleteMessage(arg0 *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteMessage", arg0)
	ret0, _ := ret[0].(*sqs.DeleteMessageOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteMessage indicates an expected call of DeleteMessage.
func (mr *MockQueueClientMockRecorder) DeleteMessage(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMessage", reflect.TypeOf((*MockQueueClient)(nil).DeleteMessage), arg0)
}

// ReceiveMessage mocks base method.
func (m *MockQueueClient) ReceiveMessage(arg0 *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReceiveMessage", arg0)
	ret0, _ := ret[0].(*sqs.ReceiveMessageOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReceiveMessage indicates an expected call of ReceiveMessage.
func (mr *MockQueueClientMockRecorder) ReceiveMessage(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReceiveMessage", reflect.TypeOf((*MockQueueClient)(nil).ReceiveMessage), arg0)
}
