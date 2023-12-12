// Code generated by MockGen. DO NOT EDIT.
// Source: database/sql/driver (interfaces: Driver,Connector,Conn,DriverContext)

// Package mock_driver is a generated GoMock package.
package mock_driver

import (
	context "context"
	driver "database/sql/driver"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockDriver is a mock of Driver interface.
type MockDriver struct {
	ctrl     *gomock.Controller
	recorder *MockDriverMockRecorder
}

// MockDriverMockRecorder is the mock recorder for MockDriver.
type MockDriverMockRecorder struct {
	mock *MockDriver
}

// NewMockDriver creates a new mock instance.
func NewMockDriver(ctrl *gomock.Controller) *MockDriver {
	mock := &MockDriver{ctrl: ctrl}
	mock.recorder = &MockDriverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDriver) EXPECT() *MockDriverMockRecorder {
	return m.recorder
}

// Open mocks base method.
func (m *MockDriver) Open(arg0 string) (driver.Conn, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Open", arg0)
	ret0, _ := ret[0].(driver.Conn)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Open indicates an expected call of Open.
func (mr *MockDriverMockRecorder) Open(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Open", reflect.TypeOf((*MockDriver)(nil).Open), arg0)
}

// MockConnector is a mock of Connector interface.
type MockConnector struct {
	ctrl     *gomock.Controller
	recorder *MockConnectorMockRecorder
}

// MockConnectorMockRecorder is the mock recorder for MockConnector.
type MockConnectorMockRecorder struct {
	mock *MockConnector
}

// NewMockConnector creates a new mock instance.
func NewMockConnector(ctrl *gomock.Controller) *MockConnector {
	mock := &MockConnector{ctrl: ctrl}
	mock.recorder = &MockConnectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConnector) EXPECT() *MockConnectorMockRecorder {
	return m.recorder
}

// Connect mocks base method.
func (m *MockConnector) Connect(arg0 context.Context) (driver.Conn, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connect", arg0)
	ret0, _ := ret[0].(driver.Conn)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Connect indicates an expected call of Connect.
func (mr *MockConnectorMockRecorder) Connect(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockConnector)(nil).Connect), arg0)
}

// Driver mocks base method.
func (m *MockConnector) Driver() driver.Driver {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Driver")
	ret0, _ := ret[0].(driver.Driver)
	return ret0
}

// Driver indicates an expected call of Driver.
func (mr *MockConnectorMockRecorder) Driver() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Driver", reflect.TypeOf((*MockConnector)(nil).Driver))
}

// MockConn is a mock of Conn interface.
type MockConn struct {
	ctrl     *gomock.Controller
	recorder *MockConnMockRecorder
}

// MockConnMockRecorder is the mock recorder for MockConn.
type MockConnMockRecorder struct {
	mock *MockConn
}

// NewMockConn creates a new mock instance.
func NewMockConn(ctrl *gomock.Controller) *MockConn {
	mock := &MockConn{ctrl: ctrl}
	mock.recorder = &MockConnMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConn) EXPECT() *MockConnMockRecorder {
	return m.recorder
}

// Begin mocks base method.
func (m *MockConn) Begin() (driver.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Begin")
	ret0, _ := ret[0].(driver.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Begin indicates an expected call of Begin.
func (mr *MockConnMockRecorder) Begin() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Begin", reflect.TypeOf((*MockConn)(nil).Begin))
}

// Close mocks base method.
func (m *MockConn) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockConnMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockConn)(nil).Close))
}

// Prepare mocks base method.
func (m *MockConn) Prepare(arg0 string) (driver.Stmt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Prepare", arg0)
	ret0, _ := ret[0].(driver.Stmt)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Prepare indicates an expected call of Prepare.
func (mr *MockConnMockRecorder) Prepare(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Prepare", reflect.TypeOf((*MockConn)(nil).Prepare), arg0)
}

// MockDriverContext is a mock of DriverContext interface.
type MockDriverContext struct {
	ctrl     *gomock.Controller
	recorder *MockDriverContextMockRecorder
}

// MockDriverContextMockRecorder is the mock recorder for MockDriverContext.
type MockDriverContextMockRecorder struct {
	mock *MockDriverContext
}

// NewMockDriverContext creates a new mock instance.
func NewMockDriverContext(ctrl *gomock.Controller) *MockDriverContext {
	mock := &MockDriverContext{ctrl: ctrl}
	mock.recorder = &MockDriverContextMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDriverContext) EXPECT() *MockDriverContextMockRecorder {
	return m.recorder
}

// OpenConnector mocks base method.
func (m *MockDriverContext) OpenConnector(arg0 string) (driver.Connector, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OpenConnector", arg0)
	ret0, _ := ret[0].(driver.Connector)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OpenConnector indicates an expected call of OpenConnector.
func (mr *MockDriverContextMockRecorder) OpenConnector(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OpenConnector", reflect.TypeOf((*MockDriverContext)(nil).OpenConnector), arg0)
}
