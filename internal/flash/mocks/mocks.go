// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aospalliance/device-flasher/internal/flash (interfaces: FactoryImageFlasher,PlatformToolsFlasher,ADBFlasher,FastbootFlasher)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	device "github.com/aospalliance/device-flasher/internal/device"
	platformtools "github.com/aospalliance/device-flasher/internal/platformtools"
	fastboot "github.com/aospalliance/device-flasher/internal/platformtools/fastboot"
	reflect "reflect"
)

// MockFactoryImageFlasher is a mock of FactoryImageFlasher interface
type MockFactoryImageFlasher struct {
	ctrl     *gomock.Controller
	recorder *MockFactoryImageFlasherMockRecorder
}

// MockFactoryImageFlasherMockRecorder is the mock recorder for MockFactoryImageFlasher
type MockFactoryImageFlasherMockRecorder struct {
	mock *MockFactoryImageFlasher
}

// NewMockFactoryImageFlasher creates a new mock instance
func NewMockFactoryImageFlasher(ctrl *gomock.Controller) *MockFactoryImageFlasher {
	mock := &MockFactoryImageFlasher{ctrl: ctrl}
	mock.recorder = &MockFactoryImageFlasherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFactoryImageFlasher) EXPECT() *MockFactoryImageFlasherMockRecorder {
	return m.recorder
}

// FlashAll mocks base method
func (m *MockFactoryImageFlasher) FlashAll(arg0 *device.Device, arg1 platformtools.PlatformToolsPath) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FlashAll", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// FlashAll indicates an expected call of FlashAll
func (mr *MockFactoryImageFlasherMockRecorder) FlashAll(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FlashAll", reflect.TypeOf((*MockFactoryImageFlasher)(nil).FlashAll), arg0, arg1)
}

// Validate mocks base method
func (m *MockFactoryImageFlasher) Validate(arg0 device.Codename) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validate", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Validate indicates an expected call of Validate
func (mr *MockFactoryImageFlasherMockRecorder) Validate(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validate", reflect.TypeOf((*MockFactoryImageFlasher)(nil).Validate), arg0)
}

// MockPlatformToolsFlasher is a mock of PlatformToolsFlasher interface
type MockPlatformToolsFlasher struct {
	ctrl     *gomock.Controller
	recorder *MockPlatformToolsFlasherMockRecorder
}

// MockPlatformToolsFlasherMockRecorder is the mock recorder for MockPlatformToolsFlasher
type MockPlatformToolsFlasherMockRecorder struct {
	mock *MockPlatformToolsFlasher
}

// NewMockPlatformToolsFlasher creates a new mock instance
func NewMockPlatformToolsFlasher(ctrl *gomock.Controller) *MockPlatformToolsFlasher {
	mock := &MockPlatformToolsFlasher{ctrl: ctrl}
	mock.recorder = &MockPlatformToolsFlasherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPlatformToolsFlasher) EXPECT() *MockPlatformToolsFlasherMockRecorder {
	return m.recorder
}

// Path mocks base method
func (m *MockPlatformToolsFlasher) Path() platformtools.PlatformToolsPath {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Path")
	ret0, _ := ret[0].(platformtools.PlatformToolsPath)
	return ret0
}

// Path indicates an expected call of Path
func (mr *MockPlatformToolsFlasherMockRecorder) Path() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Path", reflect.TypeOf((*MockPlatformToolsFlasher)(nil).Path))
}

// MockADBFlasher is a mock of ADBFlasher interface
type MockADBFlasher struct {
	ctrl     *gomock.Controller
	recorder *MockADBFlasherMockRecorder
}

// MockADBFlasherMockRecorder is the mock recorder for MockADBFlasher
type MockADBFlasherMockRecorder struct {
	mock *MockADBFlasher
}

// NewMockADBFlasher creates a new mock instance
func NewMockADBFlasher(ctrl *gomock.Controller) *MockADBFlasher {
	mock := &MockADBFlasher{ctrl: ctrl}
	mock.recorder = &MockADBFlasherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockADBFlasher) EXPECT() *MockADBFlasherMockRecorder {
	return m.recorder
}

// KillServer mocks base method
func (m *MockADBFlasher) KillServer() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "KillServer")
	ret0, _ := ret[0].(error)
	return ret0
}

// KillServer indicates an expected call of KillServer
func (mr *MockADBFlasherMockRecorder) KillServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KillServer", reflect.TypeOf((*MockADBFlasher)(nil).KillServer))
}

// RebootIntoBootloader mocks base method
func (m *MockADBFlasher) RebootIntoBootloader(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RebootIntoBootloader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RebootIntoBootloader indicates an expected call of RebootIntoBootloader
func (mr *MockADBFlasherMockRecorder) RebootIntoBootloader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RebootIntoBootloader", reflect.TypeOf((*MockADBFlasher)(nil).RebootIntoBootloader), arg0)
}

// MockFastbootFlasher is a mock of FastbootFlasher interface
type MockFastbootFlasher struct {
	ctrl     *gomock.Controller
	recorder *MockFastbootFlasherMockRecorder
}

// MockFastbootFlasherMockRecorder is the mock recorder for MockFastbootFlasher
type MockFastbootFlasherMockRecorder struct {
	mock *MockFastbootFlasher
}

// NewMockFastbootFlasher creates a new mock instance
func NewMockFastbootFlasher(ctrl *gomock.Controller) *MockFastbootFlasher {
	mock := &MockFastbootFlasher{ctrl: ctrl}
	mock.recorder = &MockFastbootFlasherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFastbootFlasher) EXPECT() *MockFastbootFlasherMockRecorder {
	return m.recorder
}

// GetBootloaderLockStatus mocks base method
func (m *MockFastbootFlasher) GetBootloaderLockStatus(arg0 string) (fastboot.FastbootLockStatus, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBootloaderLockStatus", arg0)
	ret0, _ := ret[0].(fastboot.FastbootLockStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBootloaderLockStatus indicates an expected call of GetBootloaderLockStatus
func (mr *MockFastbootFlasherMockRecorder) GetBootloaderLockStatus(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBootloaderLockStatus", reflect.TypeOf((*MockFastbootFlasher)(nil).GetBootloaderLockStatus), arg0)
}

// LockBootloader mocks base method
func (m *MockFastbootFlasher) LockBootloader(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LockBootloader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// LockBootloader indicates an expected call of LockBootloader
func (mr *MockFastbootFlasherMockRecorder) LockBootloader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LockBootloader", reflect.TypeOf((*MockFastbootFlasher)(nil).LockBootloader), arg0)
}

// Reboot mocks base method
func (m *MockFastbootFlasher) Reboot(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reboot", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reboot indicates an expected call of Reboot
func (mr *MockFastbootFlasherMockRecorder) Reboot(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reboot", reflect.TypeOf((*MockFastbootFlasher)(nil).Reboot), arg0)
}

// UnlockBootloader mocks base method
func (m *MockFastbootFlasher) UnlockBootloader(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnlockBootloader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UnlockBootloader indicates an expected call of UnlockBootloader
func (mr *MockFastbootFlasherMockRecorder) UnlockBootloader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnlockBootloader", reflect.TypeOf((*MockFastbootFlasher)(nil).UnlockBootloader), arg0)
}
