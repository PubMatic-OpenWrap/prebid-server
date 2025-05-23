// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/PubMatic-OpenWrap/prebid-server/v3/modules/pubmatic/openwrap/database (interfaces: Database)

// Package mock_database is a generated GoMock package.
package mock_database

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	adpodconfig "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adpodconfig"
	adunitconfig "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adunitconfig"
)

// MockDatabase is a mock of Database interface.
type MockDatabase struct {
	ctrl     *gomock.Controller
	recorder *MockDatabaseMockRecorder
}

// MockDatabaseMockRecorder is the mock recorder for MockDatabase.
type MockDatabaseMockRecorder struct {
	mock *MockDatabase
}

// NewMockDatabase creates a new mock instance.
func NewMockDatabase(ctrl *gomock.Controller) *MockDatabase {
	mock := &MockDatabase{ctrl: ctrl}
	mock.recorder = &MockDatabaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDatabase) EXPECT() *MockDatabaseMockRecorder {
	return m.recorder
}

// GetActivePartnerConfigurations mocks base method.
func (m *MockDatabase) GetActivePartnerConfigurations(arg0, arg1, arg2 int) (map[int]map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetActivePartnerConfigurations", arg0, arg1, arg2)
	ret0, _ := ret[0].(map[int]map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetActivePartnerConfigurations indicates an expected call of GetActivePartnerConfigurations.
func (mr *MockDatabaseMockRecorder) GetActivePartnerConfigurations(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetActivePartnerConfigurations", reflect.TypeOf((*MockDatabase)(nil).GetActivePartnerConfigurations), arg0, arg1, arg2)
}

// GetAdpodConfig mocks base method.
func (m *MockDatabase) GetAdpodConfig(arg0, arg1, arg2 int) (*adpodconfig.AdpodConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAdpodConfig", arg0, arg1, arg2)
	ret0, _ := ret[0].(*adpodconfig.AdpodConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAdpodConfig indicates an expected call of GetAdpodConfig.
func (mr *MockDatabaseMockRecorder) GetAdpodConfig(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAdpodConfig", reflect.TypeOf((*MockDatabase)(nil).GetAdpodConfig), arg0, arg1, arg2)
}

// GetAdunitConfig mocks base method.
func (m *MockDatabase) GetAdunitConfig(arg0, arg1 int) (*adunitconfig.AdUnitConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAdunitConfig", arg0, arg1)
	ret0, _ := ret[0].(*adunitconfig.AdUnitConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAdunitConfig indicates an expected call of GetAdunitConfig.
func (mr *MockDatabaseMockRecorder) GetAdunitConfig(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAdunitConfig", reflect.TypeOf((*MockDatabase)(nil).GetAdunitConfig), arg0, arg1)
}

// GetAppIntegrationPaths mocks base method.
func (m *MockDatabase) GetAppIntegrationPaths() (map[string]int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAppIntegrationPaths")
	ret0, _ := ret[0].(map[string]int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAppIntegrationPaths indicates an expected call of GetAppIntegrationPaths.
func (mr *MockDatabaseMockRecorder) GetAppIntegrationPaths() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAppIntegrationPaths", reflect.TypeOf((*MockDatabase)(nil).GetAppIntegrationPaths))
}

// GetAppSubIntegrationPaths mocks base method.
func (m *MockDatabase) GetAppSubIntegrationPaths() (map[string]int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAppSubIntegrationPaths")
	ret0, _ := ret[0].(map[string]int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAppSubIntegrationPaths indicates an expected call of GetAppSubIntegrationPaths.
func (mr *MockDatabaseMockRecorder) GetAppSubIntegrationPaths() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAppSubIntegrationPaths", reflect.TypeOf((*MockDatabase)(nil).GetAppSubIntegrationPaths))
}

// GetFSCThresholdPerDSP mocks base method.
func (m *MockDatabase) GetFSCThresholdPerDSP() (map[int]int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFSCThresholdPerDSP")
	ret0, _ := ret[0].(map[int]int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFSCThresholdPerDSP indicates an expected call of GetFSCThresholdPerDSP.
func (mr *MockDatabaseMockRecorder) GetFSCThresholdPerDSP() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFSCThresholdPerDSP", reflect.TypeOf((*MockDatabase)(nil).GetFSCThresholdPerDSP))
}

// GetGDPRCountryCodes mocks base method.
func (m *MockDatabase) GetGDPRCountryCodes() (models.HashSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGDPRCountryCodes")
	ret0, _ := ret[0].(models.HashSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGDPRCountryCodes indicates an expected call of GetGDPRCountryCodes.
func (mr *MockDatabaseMockRecorder) GetGDPRCountryCodes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGDPRCountryCodes", reflect.TypeOf((*MockDatabase)(nil).GetGDPRCountryCodes))
}

// GetMappings mocks base method.
func (m *MockDatabase) GetMappings(arg0 string, arg1 map[string]models.SlotMapping) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMappings", arg0, arg1)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMappings indicates an expected call of GetMappings.
func (mr *MockDatabaseMockRecorder) GetMappings(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMappings", reflect.TypeOf((*MockDatabase)(nil).GetMappings), arg0, arg1)
}

// GetProfileAdUnitMultiFloors mocks base method.
func (m *MockDatabase) GetProfileAdUnitMultiFloors() (models.ProfileAdUnitMultiFloors, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProfileAdUnitMultiFloors")
	ret0, _ := ret[0].(models.ProfileAdUnitMultiFloors)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProfileAdUnitMultiFloors indicates an expected call of GetProfileAdUnitMultiFloors.
func (mr *MockDatabaseMockRecorder) GetProfileAdUnitMultiFloors() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProfileAdUnitMultiFloors", reflect.TypeOf((*MockDatabase)(nil).GetProfileAdUnitMultiFloors))
}

// GetProfileTypePlatforms mocks base method.
func (m *MockDatabase) GetProfileTypePlatforms() (map[string]int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProfileTypePlatforms")
	ret0, _ := ret[0].(map[string]int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProfileTypePlatforms indicates an expected call of GetProfileTypePlatforms.
func (mr *MockDatabaseMockRecorder) GetProfileTypePlatforms() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProfileTypePlatforms", reflect.TypeOf((*MockDatabase)(nil).GetProfileTypePlatforms))
}

// GetPublisherFeatureMap mocks base method.
func (m *MockDatabase) GetPublisherFeatureMap() (map[int]map[int]models.FeatureData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPublisherFeatureMap")
	ret0, _ := ret[0].(map[int]map[int]models.FeatureData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPublisherFeatureMap indicates an expected call of GetPublisherFeatureMap.
func (mr *MockDatabaseMockRecorder) GetPublisherFeatureMap() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPublisherFeatureMap", reflect.TypeOf((*MockDatabase)(nil).GetPublisherFeatureMap))
}

// GetPublisherSlotNameHash mocks base method.
func (m *MockDatabase) GetPublisherSlotNameHash(arg0 int) (map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPublisherSlotNameHash", arg0)
	ret0, _ := ret[0].(map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPublisherSlotNameHash indicates an expected call of GetPublisherSlotNameHash.
func (mr *MockDatabaseMockRecorder) GetPublisherSlotNameHash(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPublisherSlotNameHash", reflect.TypeOf((*MockDatabase)(nil).GetPublisherSlotNameHash), arg0)
}

// GetPublisherVASTTags mocks base method.
func (m *MockDatabase) GetPublisherVASTTags(arg0 int) (map[int]*models.VASTTag, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPublisherVASTTags", arg0)
	ret0, _ := ret[0].(map[int]*models.VASTTag)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPublisherVASTTags indicates an expected call of GetPublisherVASTTags.
func (mr *MockDatabaseMockRecorder) GetPublisherVASTTags(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPublisherVASTTags", reflect.TypeOf((*MockDatabase)(nil).GetPublisherVASTTags), arg0)
}

// GetWrapperSlotMappings mocks base method.
func (m *MockDatabase) GetWrapperSlotMappings(arg0 map[int]map[string]string, arg1, arg2 int) (map[int][]models.SlotMapping, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWrapperSlotMappings", arg0, arg1, arg2)
	ret0, _ := ret[0].(map[int][]models.SlotMapping)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWrapperSlotMappings indicates an expected call of GetWrapperSlotMappings.
func (mr *MockDatabaseMockRecorder) GetWrapperSlotMappings(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWrapperSlotMappings", reflect.TypeOf((*MockDatabase)(nil).GetWrapperSlotMappings), arg0, arg1, arg2)
}
