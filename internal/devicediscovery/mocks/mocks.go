// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aospalliance/device-flasher/internal/devicediscovery (interfaces: DeviceDiscoverer)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockDeviceDiscoverer is a mock of DeviceDiscoverer interface
type MockDeviceDiscoverer struct {
	ctrl     *gomock.Controller
	recorder *MockDeviceDiscovererMockRecorder
}

// MockDeviceDiscovererMockRecorder is the mock recorder for MockDeviceDiscoverer
type MockDeviceDiscovererMockRecorder struct {
	mock *MockDeviceDiscoverer
}

// NewMockDeviceDiscoverer creates a new mock instance
func NewMockDeviceDiscoverer(ctrl *gomock.Controller) *MockDeviceDiscoverer {
	mock := &MockDeviceDiscoverer{ctrl: ctrl}
	mock.recorder = &MockDeviceDiscovererMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDeviceDiscoverer) EXPECT() *MockDeviceDiscovererMockRecorder {
	return m.recorder
}

// GetDeviceCodename mocks base method
func (m *MockDeviceDiscoverer) GetDeviceCodename(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeviceCodename", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeviceCodename indicates an expected call of GetDeviceCodename
func (mr *MockDeviceDiscovererMockRecorder) GetDeviceCodename(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeviceCodename", reflect.TypeOf((*MockDeviceDiscoverer)(nil).GetDeviceCodename), arg0)
}

// GetDeviceIds mocks base method
func (m *MockDeviceDiscoverer) GetDeviceIds() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeviceIds")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeviceIds indicates an expected call of GetDeviceIds
func (mr *MockDeviceDiscovererMockRecorder) GetDeviceIds() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeviceIds", reflect.TypeOf((*MockDeviceDiscoverer)(nil).GetDeviceIds))
}

// Name mocks base method
func (m *MockDeviceDiscoverer) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name
func (mr *MockDeviceDiscovererMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockDeviceDiscoverer)(nil).Name))
}
