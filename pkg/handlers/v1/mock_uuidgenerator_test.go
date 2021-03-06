// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/asecurityteam/ipam-facade/pkg/domain (interfaces: UUIDGenerator)

// Package v1 is a generated GoMock package.
package v1

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockUUIDGenerator is a mock of UUIDGenerator interface
type MockUUIDGenerator struct {
	ctrl     *gomock.Controller
	recorder *MockUUIDGeneratorMockRecorder
}

// MockUUIDGeneratorMockRecorder is the mock recorder for MockUUIDGenerator
type MockUUIDGeneratorMockRecorder struct {
	mock *MockUUIDGenerator
}

// NewMockUUIDGenerator creates a new mock instance
func NewMockUUIDGenerator(ctrl *gomock.Controller) *MockUUIDGenerator {
	mock := &MockUUIDGenerator{ctrl: ctrl}
	mock.recorder = &MockUUIDGeneratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockUUIDGenerator) EXPECT() *MockUUIDGeneratorMockRecorder {
	return m.recorder
}

// NewUUIDString mocks base method
func (m *MockUUIDGenerator) NewUUIDString() (string, error) {
	ret := m.ctrl.Call(m, "NewUUIDString")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewUUIDString indicates an expected call of NewUUIDString
func (mr *MockUUIDGeneratorMockRecorder) NewUUIDString() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewUUIDString", reflect.TypeOf((*MockUUIDGenerator)(nil).NewUUIDString))
}
