// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/PubMatic-OpenWrap/prebid-server/v2/analytics/pubmatic/mhttp (interfaces: HttpCallInterface,MultiHttpContextInterface)

// Package mock_mhttp is a generated GoMock package.
package mock_mhttp

import (
	reflect "reflect"
	sync "sync"

	gomock "github.com/golang/mock/gomock"
	mhttp "github.com/prebid/prebid-server/v2/analytics/pubmatic/mhttp"
)

// MockHttpCallInterface is a mock of HttpCallInterface interface.
type MockHttpCallInterface struct {
	ctrl     *gomock.Controller
	recorder *MockHttpCallInterfaceMockRecorder
}

// MockHttpCallInterfaceMockRecorder is the mock recorder for MockHttpCallInterface.
type MockHttpCallInterfaceMockRecorder struct {
	mock *MockHttpCallInterface
}

// NewMockHttpCallInterface creates a new mock instance.
func NewMockHttpCallInterface(ctrl *gomock.Controller) *MockHttpCallInterface {
	mock := &MockHttpCallInterface{ctrl: ctrl}
	mock.recorder = &MockHttpCallInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHttpCallInterface) EXPECT() *MockHttpCallInterfaceMockRecorder {
	return m.recorder
}

// AddCookie mocks base method.
func (m *MockHttpCallInterface) AddCookie(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddCookie", arg0, arg1)
}

// AddCookie indicates an expected call of AddCookie.
func (mr *MockHttpCallInterfaceMockRecorder) AddCookie(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddCookie", reflect.TypeOf((*MockHttpCallInterface)(nil).AddCookie), arg0, arg1)
}

// AddHeader mocks base method.
func (m *MockHttpCallInterface) AddHeader(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddHeader", arg0, arg1)
}

// AddHeader indicates an expected call of AddHeader.
func (mr *MockHttpCallInterfaceMockRecorder) AddHeader(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddHeader", reflect.TypeOf((*MockHttpCallInterface)(nil).AddHeader), arg0, arg1)
}

// GetResponseBody mocks base method.
func (m *MockHttpCallInterface) GetResponseBody() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetResponseBody")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetResponseBody indicates an expected call of GetResponseBody.
func (mr *MockHttpCallInterfaceMockRecorder) GetResponseBody() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResponseBody", reflect.TypeOf((*MockHttpCallInterface)(nil).GetResponseBody))
}

// getError mocks base method.
func (m *MockHttpCallInterface) getError() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "getError")
	ret0, _ := ret[0].(error)
	return ret0
}

// getError indicates an expected call of getError.
func (mr *MockHttpCallInterfaceMockRecorder) getError() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "getError", reflect.TypeOf((*MockHttpCallInterface)(nil).getError))
}

// submit mocks base method.
func (m *MockHttpCallInterface) submit(arg0 *sync.WaitGroup) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "submit", arg0)
}

// submit indicates an expected call of submit.
func (mr *MockHttpCallInterfaceMockRecorder) submit(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "submit", reflect.TypeOf((*MockHttpCallInterface)(nil).submit), arg0)
}

// MockMultiHttpContextInterface is a mock of MultiHttpContextInterface interface.
type MockMultiHttpContextInterface struct {
	ctrl     *gomock.Controller
	recorder *MockMultiHttpContextInterfaceMockRecorder
}

// MockMultiHttpContextInterfaceMockRecorder is the mock recorder for MockMultiHttpContextInterface.
type MockMultiHttpContextInterfaceMockRecorder struct {
	mock *MockMultiHttpContextInterface
}

// NewMockMultiHttpContextInterface creates a new mock instance.
func NewMockMultiHttpContextInterface(ctrl *gomock.Controller) *MockMultiHttpContextInterface {
	mock := &MockMultiHttpContextInterface{ctrl: ctrl}
	mock.recorder = &MockMultiHttpContextInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMultiHttpContextInterface) EXPECT() *MockMultiHttpContextInterfaceMockRecorder {
	return m.recorder
}

// AddHttpCall mocks base method.
func (m *MockMultiHttpContextInterface) AddHttpCall(arg0 mhttp.HttpCallInterface) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddHttpCall", arg0)
}

// AddHttpCall indicates an expected call of AddHttpCall.
func (mr *MockMultiHttpContextInterfaceMockRecorder) AddHttpCall(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddHttpCall", reflect.TypeOf((*MockMultiHttpContextInterface)(nil).AddHttpCall), arg0)
}

// Execute mocks base method.
func (m *MockMultiHttpContextInterface) Execute() (int, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute")
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// Execute indicates an expected call of Execute.
func (mr *MockMultiHttpContextInterfaceMockRecorder) Execute() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockMultiHttpContextInterface)(nil).Execute))
}
