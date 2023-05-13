// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/services/scheduler_service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/didil/inhooks/pkg/models"
	gomock "github.com/golang/mock/gomock"
)

// MockSchedulerService is a mock of SchedulerService interface.
type MockSchedulerService struct {
	ctrl     *gomock.Controller
	recorder *MockSchedulerServiceMockRecorder
}

// MockSchedulerServiceMockRecorder is the mock recorder for MockSchedulerService.
type MockSchedulerServiceMockRecorder struct {
	mock *MockSchedulerService
}

// NewMockSchedulerService creates a new mock instance.
func NewMockSchedulerService(ctrl *gomock.Controller) *MockSchedulerService {
	mock := &MockSchedulerService{ctrl: ctrl}
	mock.recorder = &MockSchedulerServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSchedulerService) EXPECT() *MockSchedulerServiceMockRecorder {
	return m.recorder
}

// MoveDueScheduled mocks base method.
func (m *MockSchedulerService) MoveDueScheduled(ctx context.Context, f *models.Flow, sink *models.Sink) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MoveDueScheduled", ctx, f, sink)
	ret0, _ := ret[0].(error)
	return ret0
}

// MoveDueScheduled indicates an expected call of MoveDueScheduled.
func (mr *MockSchedulerServiceMockRecorder) MoveDueScheduled(ctx, f, sink interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MoveDueScheduled", reflect.TypeOf((*MockSchedulerService)(nil).MoveDueScheduled), ctx, f, sink)
}
