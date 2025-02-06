// Code generated by MockGen. DO NOT EDIT.
// Source: CutMe/internal/application/service (interfaces: Notifier)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockNotifier is a mock of Notifier interface.
type MockNotifier struct {
	ctrl     *gomock.Controller
	recorder *MockNotifierMockRecorder
}

// MockNotifierMockRecorder is the mock recorder for MockNotifier.
type MockNotifierMockRecorder struct {
	mock *MockNotifier
}

// NewMockNotifier creates a new mock instance.
func NewMockNotifier(ctrl *gomock.Controller) *MockNotifier {
	mock := &MockNotifier{ctrl: ctrl}
	mock.recorder = &MockNotifierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNotifier) EXPECT() *MockNotifierMockRecorder {
	return m.recorder
}

// SendFailureEmail mocks base method.
func (m *MockNotifier) SendFailureEmail(arg0, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendFailureEmail", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendFailureEmail indicates an expected call of SendFailureEmail.
func (mr *MockNotifierMockRecorder) SendFailureEmail(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendFailureEmail", reflect.TypeOf((*MockNotifier)(nil).SendFailureEmail), arg0, arg1, arg2)
}

// SendSuccessEmailWithLinks mocks base method.
func (m *MockNotifier) SendSuccessEmailWithLinks(arg0, arg1, arg2, arg3 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendSuccessEmailWithLinks", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendSuccessEmailWithLinks indicates an expected call of SendSuccessEmailWithLinks.
func (mr *MockNotifierMockRecorder) SendSuccessEmailWithLinks(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendSuccessEmailWithLinks", reflect.TypeOf((*MockNotifier)(nil).SendSuccessEmailWithLinks), arg0, arg1, arg2, arg3)
}
