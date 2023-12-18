// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/geodb (interfaces: Geography)

// Package mock_geodb is a generated GoMock package.
package mock_geodb

import (
	gomock "github.com/golang/mock/gomock"
	geodb "github.com/prebid/prebid-server/modules/pubmatic/openwrap/geodb"
	reflect "reflect"
)

// MockGeography is a mock of Geography interface
type MockGeography struct {
	ctrl     *gomock.Controller
	recorder *MockGeographyMockRecorder
}

// MockGeographyMockRecorder is the mock recorder for MockGeography
type MockGeographyMockRecorder struct {
	mock *MockGeography
}

// NewMockGeography creates a new mock instance
func NewMockGeography(ctrl *gomock.Controller) *MockGeography {
	mock := &MockGeography{ctrl: ctrl}
	mock.recorder = &MockGeographyMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockGeography) EXPECT() *MockGeographyMockRecorder {
	return m.recorder
}

// InitGeoDBClient mocks base method
func (m *MockGeography) InitGeoDBClient(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitGeoDBClient", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// InitGeoDBClient indicates an expected call of InitGeoDBClient
func (mr *MockGeographyMockRecorder) InitGeoDBClient(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitGeoDBClient", reflect.TypeOf((*MockGeography)(nil).InitGeoDBClient), arg0)
}

// LookUp mocks base method
func (m *MockGeography) LookUp(arg0 string) (*geodb.GeoInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LookUp", arg0)
	ret0, _ := ret[0].(*geodb.GeoInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LookUp indicates an expected call of LookUp
func (mr *MockGeographyMockRecorder) LookUp(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LookUp", reflect.TypeOf((*MockGeography)(nil).LookUp), arg0)
}
