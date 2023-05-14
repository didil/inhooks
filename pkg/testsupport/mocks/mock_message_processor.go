// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/services/message_processor.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/didil/inhooks/pkg/models"
	gomock "github.com/golang/mock/gomock"
)

// MockMessageProcessor is a mock of MessageProcessor interface.
type MockMessageProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockMessageProcessorMockRecorder
}

// MockMessageProcessorMockRecorder is the mock recorder for MockMessageProcessor.
type MockMessageProcessorMockRecorder struct {
	mock *MockMessageProcessor
}

// NewMockMessageProcessor creates a new mock instance.
func NewMockMessageProcessor(ctrl *gomock.Controller) *MockMessageProcessor {
	mock := &MockMessageProcessor{ctrl: ctrl}
	mock.recorder = &MockMessageProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMessageProcessor) EXPECT() *MockMessageProcessorMockRecorder {
	return m.recorder
}

// Process mocks base method.
func (m_2 *MockMessageProcessor) Process(ctx context.Context, sink *models.Sink, m *models.Message) error {
	m_2.ctrl.T.Helper()
	ret := m_2.ctrl.Call(m_2, "Process", ctx, sink, m)
	ret0, _ := ret[0].(error)
	return ret0
}

// Process indicates an expected call of Process.
func (mr *MockMessageProcessorMockRecorder) Process(ctx, sink, m interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Process", reflect.TypeOf((*MockMessageProcessor)(nil).Process), ctx, sink, m)
}