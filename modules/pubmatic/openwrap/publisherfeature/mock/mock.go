// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/PubMatic-OpenWrap/prebid-server/v2/modules/pubmatic/openwrap/publisherfeature (interfaces: Feature)

// Package mock_publisherfeature is a generated GoMock package.
package mock_publisherfeature

import (
	gomock "github.com/golang/mock/gomock"
	models "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	reflect "reflect"
)

// MockFeature is a mock of Feature interface
type MockFeature struct {
	ctrl     *gomock.Controller
	recorder *MockFeatureMockRecorder
}

// MockFeatureMockRecorder is the mock recorder for MockFeature
type MockFeatureMockRecorder struct {
	mock *MockFeature
}

// NewMockFeature creates a new mock instance
func NewMockFeature(ctrl *gomock.Controller) *MockFeature {
	mock := &MockFeature{ctrl: ctrl}
	mock.recorder = &MockFeatureMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFeature) EXPECT() *MockFeatureMockRecorder {
	return m.recorder
}

// GetApplovinABTestFloors mocks base method
func (m *MockFeature) GetApplovinABTestFloors(arg0 int, arg1 string) models.ApplovinAdUnitFloors {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplovinABTestFloors", arg0, arg1)
	ret0, _ := ret[0].(models.ApplovinAdUnitFloors)
	return ret0
}

// GetApplovinABTestFloors indicates an expected call of GetApplovinABTestFloors
func (mr *MockFeatureMockRecorder) GetApplovinABTestFloors(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplovinABTestFloors", reflect.TypeOf((*MockFeature)(nil).GetApplovinABTestFloors), arg0, arg1)
}

// IsAmpMultiformatEnabled mocks base method
func (m *MockFeature) IsAmpMultiformatEnabled(arg0 int) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsAmpMultiformatEnabled", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsAmpMultiformatEnabled indicates an expected call of IsAmpMultiformatEnabled
func (mr *MockFeatureMockRecorder) IsAmpMultiformatEnabled(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAmpMultiformatEnabled", reflect.TypeOf((*MockFeature)(nil).IsAmpMultiformatEnabled), arg0)
}

// IsAnalyticsTrackingThrottled mocks base method
func (m *MockFeature) IsAnalyticsTrackingThrottled(arg0, arg1 int) (bool, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsAnalyticsTrackingThrottled", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// IsAnalyticsTrackingThrottled indicates an expected call of IsAnalyticsTrackingThrottled
func (mr *MockFeatureMockRecorder) IsAnalyticsTrackingThrottled(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAnalyticsTrackingThrottled", reflect.TypeOf((*MockFeature)(nil).IsAnalyticsTrackingThrottled), arg0, arg1)
}

// IsApplovinABTestEnabled mocks base method
func (m *MockFeature) IsApplovinABTestEnabled(arg0 int, arg1 string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsApplovinABTestEnabled", arg0, arg1)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsApplovinABTestEnabled indicates an expected call of IsApplovinABTestEnabled
func (mr *MockFeatureMockRecorder) IsApplovinABTestEnabled(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsApplovinABTestEnabled", reflect.TypeOf((*MockFeature)(nil).IsApplovinABTestEnabled), arg0, arg1)
}

// IsBidRecoveryEnabled mocks base method
func (m *MockFeature) IsBidRecoveryEnabled(arg0, arg1 int) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsBidRecoveryEnabled", arg0, arg1)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsBidRecoveryEnabled indicates an expected call of IsBidRecoveryEnabled
func (mr *MockFeatureMockRecorder) IsBidRecoveryEnabled(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsBidRecoveryEnabled", reflect.TypeOf((*MockFeature)(nil).IsBidRecoveryEnabled), arg0, arg1)
}

// IsFscApplicable mocks base method
func (m *MockFeature) IsFscApplicable(arg0 int, arg1 string, arg2 int) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsFscApplicable", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsFscApplicable indicates an expected call of IsFscApplicable
func (mr *MockFeatureMockRecorder) IsFscApplicable(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsFscApplicable", reflect.TypeOf((*MockFeature)(nil).IsFscApplicable), arg0, arg1, arg2)
}

// IsMaxFloorsEnabled mocks base method
func (m *MockFeature) IsMaxFloorsEnabled(arg0 int) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsMaxFloorsEnabled", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsMaxFloorsEnabled indicates an expected call of IsMaxFloorsEnabled
func (mr *MockFeatureMockRecorder) IsMaxFloorsEnabled(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsMaxFloorsEnabled", reflect.TypeOf((*MockFeature)(nil).IsMaxFloorsEnabled), arg0)
}

// IsTBFFeatureEnabled mocks base method
func (m *MockFeature) IsTBFFeatureEnabled(arg0, arg1 int) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsTBFFeatureEnabled", arg0, arg1)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsTBFFeatureEnabled indicates an expected call of IsTBFFeatureEnabled
func (mr *MockFeatureMockRecorder) IsTBFFeatureEnabled(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsTBFFeatureEnabled", reflect.TypeOf((*MockFeature)(nil).IsTBFFeatureEnabled), arg0, arg1)
}
