// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/asecurityteam/ipam-facade/pkg/ipamfetcher (interfaces: CustomerFetcher)

// Package ipamfetcher is a generated GoMock package.
package ipamfetcher

import (
	context "context"
	reflect "reflect"

	domain "github.com/asecurityteam/ipam-facade/pkg/domain"
	gomock "github.com/golang/mock/gomock"
)

// MockCustomerFetcher is a mock of CustomerFetcher interface
type MockCustomerFetcher struct {
	ctrl     *gomock.Controller
	recorder *MockCustomerFetcherMockRecorder
}

// MockCustomerFetcherMockRecorder is the mock recorder for MockCustomerFetcher
type MockCustomerFetcherMockRecorder struct {
	mock *MockCustomerFetcher
}

// NewMockCustomerFetcher creates a new mock instance
func NewMockCustomerFetcher(ctrl *gomock.Controller) *MockCustomerFetcher {
	mock := &MockCustomerFetcher{ctrl: ctrl}
	mock.recorder = &MockCustomerFetcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCustomerFetcher) EXPECT() *MockCustomerFetcherMockRecorder {
	return m.recorder
}

// FetchCustomers mocks base method
func (m *MockCustomerFetcher) FetchCustomers(arg0 context.Context) ([]domain.Customer, error) {
	ret := m.ctrl.Call(m, "FetchCustomers", arg0)
	ret0, _ := ret[0].([]domain.Customer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchCustomers indicates an expected call of FetchCustomers
func (mr *MockCustomerFetcherMockRecorder) FetchCustomers(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchCustomers", reflect.TypeOf((*MockCustomerFetcher)(nil).FetchCustomers), arg0)
}
