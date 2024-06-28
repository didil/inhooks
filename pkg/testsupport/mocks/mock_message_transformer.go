// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/services/message_transformer.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/didil/inhooks/pkg/models"
	gomock "github.com/golang/mock/gomock"
)

// MockMessageTransformer is a mock of MessageTransformer interface.
type MockMessageTransformer struct {
	ctrl     *gomock.Controller
	recorder *MockMessageTransformerMockRecorder
}

// MockMessageTransformerMockRecorder is the mock recorder for MockMessageTransformer.
type MockMessageTransformerMockRecorder struct {
	mock *MockMessageTransformer
}

// NewMockMessageTransformer creates a new mock instance.
func NewMockMessageTransformer(ctrl *gomock.Controller) *MockMessageTransformer {
	mock := &MockMessageTransformer{ctrl: ctrl}
	mock.recorder = &MockMessageTransformerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMessageTransformer) EXPECT() *MockMessageTransformerMockRecorder {
	return m.recorder
}

// Transform mocks base method.
func (m *MockMessageTransformer) Transform(ctx context.Context, transformDefinition *models.TransformDefinition, message *models.Message) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Transform", ctx, transformDefinition, message)
	ret0, _ := ret[0].(error)
	return ret0
}

// Transform indicates an expected call of Transform.
func (mr *MockMessageTransformerMockRecorder) Transform(ctx, transformDefinition, message interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Transform", reflect.TypeOf((*MockMessageTransformer)(nil).Transform), ctx, transformDefinition, message)
}