// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/services/retry_calculator.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockRetryCalculator is a mock of RetryCalculator interface.
type MockRetryCalculator struct {
	ctrl     *gomock.Controller
	recorder *MockRetryCalculatorMockRecorder
}

// MockRetryCalculatorMockRecorder is the mock recorder for MockRetryCalculator.
type MockRetryCalculatorMockRecorder struct {
	mock *MockRetryCalculator
}

// NewMockRetryCalculator creates a new mock instance.
func NewMockRetryCalculator(ctrl *gomock.Controller) *MockRetryCalculator {
	mock := &MockRetryCalculator{ctrl: ctrl}
	mock.recorder = &MockRetryCalculatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRetryCalculator) EXPECT() *MockRetryCalculatorMockRecorder {
	return m.recorder
}

// NextAttemptInterval mocks base method.
func (m *MockRetryCalculator) NextAttemptInterval(attemptsCount int, retryInterval *time.Duration, retryExpMultiplier *float64) time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NextAttemptInterval", attemptsCount, retryInterval, retryExpMultiplier)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// NextAttemptInterval indicates an expected call of NextAttemptInterval.
func (mr *MockRetryCalculatorMockRecorder) NextAttemptInterval(attemptsCount, retryInterval, retryExpMultiplier interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NextAttemptInterval", reflect.TypeOf((*MockRetryCalculator)(nil).NextAttemptInterval), attemptsCount, retryInterval, retryExpMultiplier)
}
