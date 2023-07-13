// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/metrics/stats (interfaces: HttpClient,WorkerPool)

// Package mock_stats is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	http "net/http"
	reflect "reflect"
)

// MockHttpClient is a mock of HttpClient interface
type MockHttpClient struct {
	ctrl     *gomock.Controller
	recorder *MockHttpClientMockRecorder
}

// MockHttpClientMockRecorder is the mock recorder for MockHttpClient
type MockHttpClientMockRecorder struct {
	mock *MockHttpClient
}

// NewMockHttpClient creates a new mock instance
func NewMockHttpClient(ctrl *gomock.Controller) *MockHttpClient {
	mock := &MockHttpClient{ctrl: ctrl}
	mock.recorder = &MockHttpClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHttpClient) EXPECT() *MockHttpClientMockRecorder {
	return m.recorder
}

// Do mocks base method
func (m *MockHttpClient) Do(arg0 *http.Request) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Do", arg0)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Do indicates an expected call of Do
func (mr *MockHttpClientMockRecorder) Do(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MockHttpClient)(nil).Do), arg0)
}

// MockWorkerPool is a mock of WorkerPool interface
type MockWorkerPool struct {
	ctrl     *gomock.Controller
	recorder *MockWorkerPoolMockRecorder
}

// MockWorkerPoolMockRecorder is the mock recorder for MockWorkerPool
type MockWorkerPoolMockRecorder struct {
	mock *MockWorkerPool
}

// NewMockWorkerPool creates a new mock instance
func NewMockWorkerPool(ctrl *gomock.Controller) *MockWorkerPool {
	mock := &MockWorkerPool{ctrl: ctrl}
	mock.recorder = &MockWorkerPoolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWorkerPool) EXPECT() *MockWorkerPoolMockRecorder {
	return m.recorder
}

// TrySubmit mocks base method
func (m *MockWorkerPool) TrySubmit(arg0 func()) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TrySubmit", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// TrySubmit indicates an expected call of TrySubmit
func (mr *MockWorkerPoolMockRecorder) TrySubmit(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TrySubmit", reflect.TypeOf((*MockWorkerPool)(nil).TrySubmit), arg0)
}