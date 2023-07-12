// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/pm-nilesh-chate/prebid-server/modules/pubmatic/openwrap/cache (interfaces: Cache)

// Package mock_cache is a generated GoMock package.
package mock_cache

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	openrtb2 "github.com/prebid/openrtb/v19/openrtb2"
	models "github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	adunitconfig "github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

// MockCache is a mock of Cache interface.
type MockCache struct {
	ctrl     *gomock.Controller
	recorder *MockCacheMockRecorder
}

// MockCacheMockRecorder is the mock recorder for MockCache.
type MockCacheMockRecorder struct {
	mock *MockCache
}

// NewMockCache creates a new mock instance.
func NewMockCache(ctrl *gomock.Controller) *MockCache {
	mock := &MockCache{ctrl: ctrl}
	mock.recorder = &MockCacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCache) EXPECT() *MockCacheMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockCache) Get(arg0 string) (interface{}, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockCacheMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCache)(nil).Get), arg0)
}

// GetAdunitConfigFromCache mocks base method.
func (m *MockCache) GetAdunitConfigFromCache(arg0 *openrtb2.BidRequest, arg1, arg2, arg3 int) *adunitconfig.AdUnitConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAdunitConfigFromCache", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*adunitconfig.AdUnitConfig)
	return ret0
}

// GetAdunitConfigFromCache indicates an expected call of GetAdunitConfigFromCache.
func (mr *MockCacheMockRecorder) GetAdunitConfigFromCache(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAdunitConfigFromCache", reflect.TypeOf((*MockCache)(nil).GetAdunitConfigFromCache), arg0, arg1, arg2, arg3)
}

// GetFSCDisabledPublishers mocks base method.
func (m *MockCache) GetFSCDisabledPublishers() (map[int]struct{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFSCDisabledPublishers")
	ret0, _ := ret[0].(map[int]struct{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFSCDisabledPublishers indicates an expected call of GetFSCDisabledPublishers.
func (mr *MockCacheMockRecorder) GetFSCDisabledPublishers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFSCDisabledPublishers", reflect.TypeOf((*MockCache)(nil).GetFSCDisabledPublishers))
}

// GetFSCThresholdPerDSP mocks base method.
func (m *MockCache) GetFSCThresholdPerDSP() (map[int]int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFSCThresholdPerDSP")
	ret0, _ := ret[0].(map[int]int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFSCThresholdPerDSP indicates an expected call of GetFSCThresholdPerDSP.
func (mr *MockCacheMockRecorder) GetFSCThresholdPerDSP() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFSCThresholdPerDSP", reflect.TypeOf((*MockCache)(nil).GetFSCThresholdPerDSP))
}

// GetMappingsFromCacheV25 mocks base method.
func (m *MockCache) GetMappingsFromCacheV25(arg0 models.RequestCtx, arg1 int) map[string]models.SlotMapping {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMappingsFromCacheV25", arg0, arg1)
	ret0, _ := ret[0].(map[string]models.SlotMapping)
	return ret0
}

// GetMappingsFromCacheV25 indicates an expected call of GetMappingsFromCacheV25.
func (mr *MockCacheMockRecorder) GetMappingsFromCacheV25(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMappingsFromCacheV25", reflect.TypeOf((*MockCache)(nil).GetMappingsFromCacheV25), arg0, arg1)
}

// GetPartnerConfigMap mocks base method.
func (m *MockCache) GetPartnerConfigMap(arg0, arg1, arg2 int) (map[int]map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPartnerConfigMap", arg0, arg1, arg2)
	ret0, _ := ret[0].(map[int]map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPartnerConfigMap indicates an expected call of GetPartnerConfigMap.
func (mr *MockCacheMockRecorder) GetPartnerConfigMap(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPartnerConfigMap", reflect.TypeOf((*MockCache)(nil).GetPartnerConfigMap), arg0, arg1, arg2)
}

// GetPublisherVASTTagsFromCache mocks base method.
func (m *MockCache) GetPublisherVASTTagsFromCache(arg0 int) map[int]*models.VASTTag {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPublisherVASTTagsFromCache", arg0)
	ret0, _ := ret[0].(map[int]*models.VASTTag)
	return ret0
}

// GetPublisherVASTTagsFromCache indicates an expected call of GetPublisherVASTTagsFromCache.
func (mr *MockCacheMockRecorder) GetPublisherVASTTagsFromCache(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPublisherVASTTagsFromCache", reflect.TypeOf((*MockCache)(nil).GetPublisherVASTTagsFromCache), arg0)
}

// GetSlotToHashValueMapFromCacheV25 mocks base method.
func (m *MockCache) GetSlotToHashValueMapFromCacheV25(arg0 models.RequestCtx, arg1 int) models.SlotMappingInfo {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSlotToHashValueMapFromCacheV25", arg0, arg1)
	ret0, _ := ret[0].(models.SlotMappingInfo)
	return ret0
}

// GetSlotToHashValueMapFromCacheV25 indicates an expected call of GetSlotToHashValueMapFromCacheV25.
func (mr *MockCacheMockRecorder) GetSlotToHashValueMapFromCacheV25(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSlotToHashValueMapFromCacheV25", reflect.TypeOf((*MockCache)(nil).GetSlotToHashValueMapFromCacheV25), arg0, arg1)
}

// Set mocks base method.
func (m *MockCache) Set(arg0 string, arg1 interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Set", arg0, arg1)
}

// Set indicates an expected call of Set.
func (mr *MockCacheMockRecorder) Set(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockCache)(nil).Set), arg0, arg1)
}
