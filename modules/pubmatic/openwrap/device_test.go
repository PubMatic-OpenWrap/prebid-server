package openwrap

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func TestPopulateDeviceExt(t *testing.T) {
	type args struct {
		device *openrtb2.Device
	}

	type want struct {
		deviceCtx models.DeviceCtx
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: `nil_request`,
			args: args{},
			want: want{
				deviceCtx: models.DeviceCtx{},
			},
		},
		{
			name: `invalid_device_ext`,
			args: args{
				device: &openrtb2.Device{
					Ext: json.RawMessage(`invalid ext`),
				},
			},
			want: want{
				deviceCtx: models.DeviceCtx{},
			},
		},
		{
			name: `ifa_present`,
			args: args{
				device: &openrtb2.Device{
					IFA: `test_ifa`,
				},
			},
			want: want{
				deviceCtx: models.DeviceCtx{
					DeviceIFA: `test_ifa`,
					ID:        "test_ifa",
				},
			},
		},
		{
			name: `ifa_type_key_absent`,
			args: args{
				device: &openrtb2.Device{
					IFA: `test_ifa`,
					Ext: json.RawMessage(`{"anykey":"anyval"}`),
				},
			},
			want: want{
				deviceCtx: models.DeviceCtx{
					DeviceIFA: `test_ifa`,
					ID:        "test_ifa",
					Ext: func() *models.ExtDevice {
						deviceExt := &models.ExtDevice{}
						deviceExt.UnmarshalJSON([]byte(`{"anykey": "anyval"}`))
						return deviceExt
					}(),
				},
			},
		},
		{
			name: `invalid_data_type_for_ifa_type_key`,
			args: args{
				device: &openrtb2.Device{
					IFA: `test_ifa`,
					Ext: json.RawMessage(`{"ifa_type": 123}`)},
			},
			want: want{
				deviceCtx: models.DeviceCtx{
					DeviceIFA: `test_ifa`,
					ID:        "test_ifa",
					Ext:       models.NewExtDevice(),
				},
			},
		},
		{
			name: `ifa_type_missing_in_DeviceIFATypeID_mapping`,
			args: args{
				device: &openrtb2.Device{
					IFA: `test_ifa`,
					Ext: json.RawMessage(`{"ifa_type": "anything"}`),
				},
			},
			want: want{
				/* removed_invalid_ifatype */
				deviceCtx: models.DeviceCtx{
					DeviceIFA: `test_ifa`,
					ID:        "test_ifa",
					Ext:       models.NewExtDevice(),
				},
			},
		},
		{
			name: `case_insensitive_ifa_type`,
			args: args{
				device: &openrtb2.Device{
					IFA: `test_ifa`,
					Ext: json.RawMessage(`{"ifa_type": "DpId"}`),
				},
			},
			want: want{
				deviceCtx: models.DeviceCtx{
					DeviceIFA: `test_ifa`,
					ID:        "test_ifa",
					IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
					Ext: func() *models.ExtDevice {
						deviceExt := &models.ExtDevice{}
						deviceExt.SetIFAType("DpId")
						return deviceExt
					}(),
				},
			},
		},
		{
			name: `valid_ifa_type`,
			args: args{
				device: &openrtb2.Device{
					IFA: `test_ifa`,
					Ext: json.RawMessage(`{"ifa_type": "sessionid"}`),
				},
			},
			want: want{
				deviceCtx: models.DeviceCtx{
					DeviceIFA: `test_ifa`,
					ID:        "test_ifa",
					IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeSESSIONID]),
					Ext: func() *models.ExtDevice {
						deviceExt := &models.ExtDevice{}
						deviceExt.SetIFAType("sessionid")
						return deviceExt
					}(),
				},
			},
		},
		{
			name: `valid_ifa_type_missing_device_ifa`,
			args: args{
				device: &openrtb2.Device{
					Ext: json.RawMessage(`{"ifa_type": "sessionid"}`),
				},
			},
			want: want{
				deviceCtx: models.DeviceCtx{
					Ext: models.NewExtDevice(),
				},
			},
		},
		{
			name: `invalid_device.ext.atts`,
			args: args{
				device: &openrtb2.Device{
					IFA: `test_ifa`,
					Ext: json.RawMessage(`{"atts": "invalid_value"}`),
				},
			},
			want: want{
				deviceCtx: models.DeviceCtx{
					DeviceIFA: `test_ifa`,
					ID:        "test_ifa",
					Ext: func() *models.ExtDevice {
						deviceExt := &models.ExtDevice{}
						deviceExt.UnmarshalJSON([]byte(`{"atts":"invalid_value"}`))
						return deviceExt
					}(),
				},
			},
		},
		{
			name: `valid_device.ext.atts`,
			args: args{
				device: &openrtb2.Device{
					Ext: json.RawMessage(`{"atts": 1}`),
				},
			},
			want: want{
				deviceCtx: models.DeviceCtx{
					Ext: func() *models.ExtDevice {
						deviceExt := &models.ExtDevice{}
						deviceExt.UnmarshalJSON([]byte(`{"atts":1}`))
						return deviceExt
					}(),
				},
			},
		},
		{
			name: `all_valid_ext_parameters`,
			args: args{
				device: &openrtb2.Device{
					IFA: `test_ifa`,
					Ext: json.RawMessage(`{"ifa_type": "sessionid","atts": 1}`),
				},
			},
			want: want{
				deviceCtx: models.DeviceCtx{
					DeviceIFA: `test_ifa`,
					ID:        "test_ifa",
					IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeSESSIONID]),
					Ext: func() *models.ExtDevice {
						deviceExt := &models.ExtDevice{}
						deviceExt.UnmarshalJSON([]byte(`{"atts":1}`))
						deviceExt.SetIFAType("sessionid")
						return deviceExt
					}(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dvc := models.DeviceCtx{}
			populateDeviceContext(&dvc, tt.args.device)
			assert.Equal(t, tt.want.deviceCtx, dvc)
		})
	}
}

func TestUpdateDeviceIFADetails(t *testing.T) {
	type args struct {
		dvc *models.DeviceCtx
	}
	tests := []struct {
		name string
		args args
		want *models.DeviceCtx
	}{
		{
			name: `empty`,
			args: args{},
			want: nil,
		},
		{
			name: `device_ext_nil`,
			args: args{
				dvc: &models.DeviceCtx{},
			},
			want: &models.DeviceCtx{},
		},
		{
			name: `device_ext_nil`,
			args: args{
				dvc: &models.DeviceCtx{},
			},
			want: &models.DeviceCtx{},
		},
		{
			name: `ifa_type_missing`,
			args: args{
				dvc: &models.DeviceCtx{
					Ext: &models.ExtDevice{},
				},
			},
			want: &models.DeviceCtx{
				Ext: &models.ExtDevice{},
			},
		},
		{
			name: `ifa_type_present_ifa_missing`,
			args: args{
				dvc: &models.DeviceCtx{
					Ext: func() *models.ExtDevice {
						ext := &models.ExtDevice{}
						ext.SetIFAType(models.DeviceIFATypeDPID)
						return ext
					}(),
				},
			},
			want: &models.DeviceCtx{
				Ext: models.NewExtDevice(),
			},
		},
		{
			name: `wrong_ifa_type`,
			args: args{
				dvc: &models.DeviceCtx{
					DeviceIFA: `sample_ifa_value`,
					Ext: func() *models.ExtDevice {
						ext := &models.ExtDevice{}
						ext.SetIFAType("wrong_ifa_type")
						return ext
					}(),
				},
			},
			want: &models.DeviceCtx{
				DeviceIFA: `sample_ifa_value`,
				Ext:       models.NewExtDevice(),
			},
		},
		{
			name: `valid_ifa_type`,
			args: args{
				dvc: &models.DeviceCtx{
					DeviceIFA: `sample_ifa_value`,
					Ext: func() *models.ExtDevice {
						ext := &models.ExtDevice{}
						ext.SetIFAType(models.DeviceIFATypeDPID)
						return ext
					}(),
				},
			},
			want: &models.DeviceCtx{
				DeviceIFA: `sample_ifa_value`,
				IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
				Ext: func() *models.ExtDevice {
					ext := &models.ExtDevice{}
					ext.SetIFAType(models.DeviceIFATypeDPID)
					return ext
				}(),
			},
		},
		{
			name: `case_insensitive_ifa_type`,
			args: args{
				dvc: &models.DeviceCtx{
					DeviceIFA: `sample_ifa_value`,
					Ext: func() *models.ExtDevice {
						ext := &models.ExtDevice{}
						ext.SetIFAType(strings.ToUpper(models.DeviceIFATypeDPID))
						return ext
					}(),
				},
			},
			want: &models.DeviceCtx{
				DeviceIFA: `sample_ifa_value`,
				IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
				Ext: func() *models.ExtDevice {
					ext := &models.ExtDevice{}
					ext.SetIFAType(strings.ToUpper(models.DeviceIFATypeDPID))
					return ext
				}(),
			},
		},
		{
			name: `ifa_type_present_session_id_present`,
			args: args{
				dvc: &models.DeviceCtx{
					Ext: func() *models.ExtDevice {
						ext := &models.ExtDevice{}
						ext.SetIFAType(strings.ToUpper(models.DeviceIFATypeDPID))
						ext.SetSessionID(`sample_session_id`)
						return ext
					}(),
				},
			},
			want: &models.DeviceCtx{
				DeviceIFA: `sample_session_id`,
				ID:        `sample_session_id`,
				IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeSESSIONID]),
				Ext: func() *models.ExtDevice {
					ext := &models.ExtDevice{}
					ext.SetIFAType(models.DeviceIFATypeSESSIONID)
					ext.SetSessionID(`sample_session_id`)
					return ext
				}(),
			},
		},
		{
			name: `ifa_type_present_session_id_missing`,
			args: args{
				dvc: &models.DeviceCtx{
					Ext: func() *models.ExtDevice {
						deviceExt := &models.ExtDevice{}
						deviceExt.SetIFAType(models.DeviceIFATypeDPID)
						return deviceExt
					}(),
				},
			},
			want: &models.DeviceCtx{
				Ext: models.NewExtDevice(),
			},
		},
		{
			name: `ifa_type_missing_session_id_present`,
			args: args{
				dvc: &models.DeviceCtx{
					DeviceIFA: `existing_ifa_id`,
					ID:        `existing_ifa_id`,
					Ext: func() *models.ExtDevice {
						deviceExt := &models.ExtDevice{}
						deviceExt.SetSessionID(`sample_session_id`)
						return deviceExt
					}(),
				},
			},
			want: &models.DeviceCtx{
				DeviceIFA: `existing_ifa_id`,
				ID:        `existing_ifa_id`,
				Ext: func() *models.ExtDevice {
					deviceExt := &models.ExtDevice{}
					deviceExt.SetSessionID(`sample_session_id`)
					return deviceExt
				}(),
			},
		},
		{
			name: `ifa_type_missing_ifa_empty_session_id_present`,
			args: args{
				dvc: &models.DeviceCtx{
					DeviceIFA: "",
					Ext: func() *models.ExtDevice {
						deviceExt := &models.ExtDevice{}
						deviceExt.SetSessionID(`sample_session_id`)
						return deviceExt
					}(),
				},
			},
			want: &models.DeviceCtx{
				IFATypeID: ptrutil.ToPtr(9),
				DeviceIFA: `sample_session_id`,
				ID:        `sample_session_id`,
				Ext: func() *models.ExtDevice {
					deviceExt := &models.ExtDevice{}
					deviceExt.SetIFAType(models.DeviceIFATypeSESSIONID)
					deviceExt.SetSessionID(`sample_session_id`)
					return deviceExt
				}(),
			},
		},
		{
			name: `ifa_type_missing_ifa_not_present_session_id_present`,
			args: args{
				dvc: &models.DeviceCtx{
					Ext: func() *models.ExtDevice {
						deviceExt := &models.ExtDevice{}
						deviceExt.SetSessionID(`sample_session_id`)
						return deviceExt
					}(),
				},
			},
			want: &models.DeviceCtx{
				IFATypeID: ptrutil.ToPtr(9),
				DeviceIFA: `sample_session_id`,
				ID:        `sample_session_id`,
				Ext: func() *models.ExtDevice {
					deviceExt := &models.ExtDevice{}
					deviceExt.SetIFAType(models.DeviceIFATypeSESSIONID)
					deviceExt.SetSessionID(`sample_session_id`)
					return deviceExt
				}(),
			},
		},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateDeviceIFADetails(tt.args.dvc)
			assert.Equal(t, tt.want, tt.args.dvc)
		})
	}
}

func TestAmendDeviceObject(t *testing.T) {
	type args struct {
		device *openrtb2.Device
		dvc    *models.DeviceCtx
	}
	tests := []struct {
		name string
		args args
		want *openrtb2.Device
	}{
		{
			name: `any_empty`,
			args: args{},
			want: nil,
		},
		{
			name: `update_ifa_details_1`,
			args: args{
				device: &openrtb2.Device{
					UA: `sample_ua`,
				},
				dvc: &models.DeviceCtx{
					DeviceIFA: `new_ifa`,
				},
			},
			want: &openrtb2.Device{
				UA:  `sample_ua`,
				IFA: `new_ifa`,
			},
		},
		{
			name: `update_ifa_details_2`,
			args: args{
				device: &openrtb2.Device{
					UA:  `sample_ua`,
					IFA: `old_ifa`,
				},
				dvc: &models.DeviceCtx{
					DeviceIFA: `new_ifa`,
				},
			},
			want: &openrtb2.Device{
				UA:  `sample_ua`,
				IFA: `new_ifa`,
			},
		},
		{
			name: `update_ext_1`,
			args: args{
				device: &openrtb2.Device{
					UA:  `sample_ua`,
					IFA: `old_ifa`,
				},
				dvc: &models.DeviceCtx{
					DeviceIFA: `new_ifa`,
					Ext: func() *models.ExtDevice {
						deviceExt := &models.ExtDevice{}
						deviceExt.SetSessionID("sample_session")
						return deviceExt
					}(),
				},
			},
			want: &openrtb2.Device{
				UA:  `sample_ua`,
				IFA: `new_ifa`,
				Ext: json.RawMessage(`{"session_id":"sample_session"}`),
			},
		},
		{
			name: `update_ext_2`,
			args: args{
				device: &openrtb2.Device{
					UA:  `sample_ua`,
					IFA: `old_ifa`,
					Ext: json.RawMessage(`{"extra_key":"missing"}`),
				},
				dvc: &models.DeviceCtx{
					DeviceIFA: `new_ifa`,
					Ext: func() *models.ExtDevice {
						deviceExt := &models.ExtDevice{}
						deviceExt.SetSessionID("sample_session")
						return deviceExt
					}(),
				},
			},
			want: &openrtb2.Device{
				UA:  `sample_ua`,
				IFA: `new_ifa`,
				Ext: json.RawMessage(`{"session_id":"sample_session"}`),
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amendDeviceObject(tt.args.device, tt.args.dvc)
			assert.Equal(t, tt.args.device, tt.want, "mismatched device object")
		})
	}
}

func TestSetDeviceIDAndModel(t *testing.T) {
	tests := []struct {
		name     string
		dvc      *models.DeviceCtx
		device   *openrtb2.Device
		expected *models.DeviceCtx
	}{
		{
			name: "DeviceIFA set",
			dvc: &models.DeviceCtx{
				DeviceIFA: "test-ifa",
			},
			device: &openrtb2.Device{
				Model: "iPhone",
				IFA:   "test-ifa",
			},
			expected: &models.DeviceCtx{
				DeviceIFA: "test-ifa",
				ID:        "test-ifa",
				Model:     "iPhone",
			},
		},
		{
			name: "DIDSHA1 set",
			dvc:  &models.DeviceCtx{},
			device: &openrtb2.Device{
				Model:   "Samsung",
				DIDSHA1: "test-didsha1",
			},
			expected: &models.DeviceCtx{
				ID:    "test-didsha1",
				Model: "Samsung",
			},
		},
		{
			name: "DIDMD5 set",
			dvc:  &models.DeviceCtx{},
			device: &openrtb2.Device{
				Model:  "Pixel",
				DIDMD5: "test-didmd5",
			},
			expected: &models.DeviceCtx{
				ID:    "test-didmd5",
				Model: "Pixel",
			},
		},
		{
			name: "DPIDSHA1 set",
			dvc:  &models.DeviceCtx{},
			device: &openrtb2.Device{
				Model:    "Huawei",
				DPIDSHA1: "test-dpidsha1",
			},
			expected: &models.DeviceCtx{
				ID:    "test-dpidsha1",
				Model: "Huawei",
			},
		},
		{
			name: "DPIDMD5 set",
			dvc:  &models.DeviceCtx{},
			device: &openrtb2.Device{
				Model:   "OnePlus",
				DPIDMD5: "test-dpidmd5",
			},
			expected: &models.DeviceCtx{
				ID:    "test-dpidmd5",
				Model: "OnePlus",
			},
		},
		{
			name: "MACSHA1 set",
			dvc:  &models.DeviceCtx{},
			device: &openrtb2.Device{
				Model:   "Xiaomi",
				MACSHA1: "test-macsha1",
			},
			expected: &models.DeviceCtx{
				ID:    "test-macsha1",
				Model: "Xiaomi",
			},
		},
		{
			name: "MACMD5 set",
			dvc:  &models.DeviceCtx{},
			device: &openrtb2.Device{
				Model:  "Oppo",
				MACMD5: "test-macmd5",
			},
			expected: &models.DeviceCtx{
				ID:    "test-macmd5",
				Model: "Oppo",
			},
		},
		{
			name: "No ID set",
			dvc:  &models.DeviceCtx{},
			device: &openrtb2.Device{
				Model: "Generic",
			},
			expected: &models.DeviceCtx{
				Model: "Generic",
			},
		},
		{
			name: "All ID set",
			dvc: &models.DeviceCtx{
				DeviceIFA: "test-ifa",
			},
			device: &openrtb2.Device{
				IFA:      "test-ifa",
				DIDSHA1:  "test-didsha1",
				DIDMD5:   "test-didmd5",
				DPIDSHA1: "test-dpidsha1",
				DPIDMD5:  "test-dpidmd5",
				MACSHA1:  "test-macsha1",
				MACMD5:   "test-macmd5",
				Model:    "iphone,11",
			},
			expected: &models.DeviceCtx{
				Model: "iphone,11",
				ID:    "test-ifa",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setDeviceIDAndModel(tt.dvc, tt.device)
			if tt.dvc.ID != tt.expected.ID {
				t.Errorf("ID = %v, want %v", tt.dvc.ID, tt.expected.ID)
			}
			if tt.dvc.Model != tt.expected.Model {
				t.Errorf("Model = %v, want %v", tt.dvc.Model, tt.expected.Model)
			}
		})
	}
}
