// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/services/message_verifier.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	models "github.com/didil/inhooks/pkg/models"
	gomock "github.com/golang/mock/gomock"
)

// MockMessageVerifier is a mock of MessageVerifier interface.
type MockMessageVerifier struct {
	ctrl     *gomock.Controller
	recorder *MockMessageVerifierMockRecorder
}

// MockMessageVerifierMockRecorder is the mock recorder for MockMessageVerifier.
type MockMessageVerifierMockRecorder struct {
	mock *MockMessageVerifier
}

// NewMockMessageVerifier creates a new mock instance.
func NewMockMessageVerifier(ctrl *gomock.Controller) *MockMessageVerifier {
	mock := &MockMessageVerifier{ctrl: ctrl}
	mock.recorder = &MockMessageVerifierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMessageVerifier) EXPECT() *MockMessageVerifierMockRecorder {
	return m.recorder
}

// Verify mocks base method.
func (m_2 *MockMessageVerifier) Verify(flow *models.Flow, m *models.Message) error {
	m_2.ctrl.T.Helper()
	ret := m_2.ctrl.Call(m_2, "Verify", flow, m)
	ret0, _ := ret[0].(error)
	return ret0
}

// Verify indicates an expected call of Verify.
func (mr *MockMessageVerifierMockRecorder) Verify(flow, m interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Verify", reflect.TypeOf((*MockMessageVerifier)(nil).Verify), flow, m)
}
