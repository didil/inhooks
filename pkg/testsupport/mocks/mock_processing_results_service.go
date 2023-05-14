// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/services/processing_results_service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/didil/inhooks/pkg/models"
	gomock "github.com/golang/mock/gomock"
)

// MockProcessingResultsService is a mock of ProcessingResultsService interface.
type MockProcessingResultsService struct {
	ctrl     *gomock.Controller
	recorder *MockProcessingResultsServiceMockRecorder
}

// MockProcessingResultsServiceMockRecorder is the mock recorder for MockProcessingResultsService.
type MockProcessingResultsServiceMockRecorder struct {
	mock *MockProcessingResultsService
}

// NewMockProcessingResultsService creates a new mock instance.
func NewMockProcessingResultsService(ctrl *gomock.Controller) *MockProcessingResultsService {
	mock := &MockProcessingResultsService{ctrl: ctrl}
	mock.recorder = &MockProcessingResultsServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProcessingResultsService) EXPECT() *MockProcessingResultsServiceMockRecorder {
	return m.recorder
}

// HandleFailed mocks base method.
func (m_2 *MockProcessingResultsService) HandleFailed(ctx context.Context, sink *models.Sink, m *models.Message, processingErr error) (models.QueueStatus, error) {
	m_2.ctrl.T.Helper()
	ret := m_2.ctrl.Call(m_2, "HandleFailed", ctx, sink, m, processingErr)
	ret0, _ := ret[0].(models.QueueStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HandleFailed indicates an expected call of HandleFailed.
func (mr *MockProcessingResultsServiceMockRecorder) HandleFailed(ctx, sink, m, processingErr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleFailed", reflect.TypeOf((*MockProcessingResultsService)(nil).HandleFailed), ctx, sink, m, processingErr)
}

// HandleOK mocks base method.
func (m_2 *MockProcessingResultsService) HandleOK(ctx context.Context, m *models.Message) error {
	m_2.ctrl.T.Helper()
	ret := m_2.ctrl.Call(m_2, "HandleOK", ctx, m)
	ret0, _ := ret[0].(error)
	return ret0
}

// HandleOK indicates an expected call of HandleOK.
func (mr *MockProcessingResultsServiceMockRecorder) HandleOK(ctx, m interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleOK", reflect.TypeOf((*MockProcessingResultsService)(nil).HandleOK), ctx, m)
}