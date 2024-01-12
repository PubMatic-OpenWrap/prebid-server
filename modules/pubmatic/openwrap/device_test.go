package openwrap

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
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
					Ext:       &models.ExtDevice{},
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
					Ext:       &models.ExtDevice{},
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
					IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
					Ext: &models.ExtDevice{
						ExtDevice: openrtb_ext.ExtDevice{
							IFAType: `DpId`,
						},
					},
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
					IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeSESSIONID]),
					Ext: &models.ExtDevice{
						ExtDevice: openrtb_ext.ExtDevice{
							IFAType: `sessionid`,
						},
					},
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
					Ext: &models.ExtDevice{},
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
					Ext: &models.ExtDevice{
						ExtDevice: openrtb_ext.ExtDevice{
							ATTS: ptrutil.ToPtr(openrtb_ext.IOSAppTrackingStatusRestricted),
						},
					},
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
					IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeSESSIONID]),
					Ext: &models.ExtDevice{
						ExtDevice: openrtb_ext.ExtDevice{
							IFAType: `sessionid`,
							ATTS:    ptrutil.ToPtr(openrtb_ext.IOSAppTrackingStatusRestricted),
						},
					},
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
					Ext: &models.ExtDevice{
						ExtDevice: openrtb_ext.ExtDevice{
							IFAType: models.DeviceIFATypeDPID,
						},
					},
				},
			},
			want: &models.DeviceCtx{
				Ext: &models.ExtDevice{},
			},
		},
		{
			name: `wrong_ifa_type`,
			args: args{
				dvc: &models.DeviceCtx{
					DeviceIFA: `sample_ifa_value`,
					Ext: &models.ExtDevice{
						ExtDevice: openrtb_ext.ExtDevice{
							IFAType: `wrong_ifa_type`,
						},
					},
				},
			},
			want: &models.DeviceCtx{
				DeviceIFA: `sample_ifa_value`,
				Ext: &models.ExtDevice{
					ExtDevice: openrtb_ext.ExtDevice{},
				},
			},
		},
		{
			name: `valid_ifa_type`,
			args: args{
				dvc: &models.DeviceCtx{
					DeviceIFA: `sample_ifa_value`,
					Ext: &models.ExtDevice{
						ExtDevice: openrtb_ext.ExtDevice{
							IFAType: models.DeviceIFATypeDPID,
						},
					},
				},
			},
			want: &models.DeviceCtx{
				DeviceIFA: `sample_ifa_value`,
				IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
				Ext: &models.ExtDevice{
					ExtDevice: openrtb_ext.ExtDevice{
						IFAType: models.DeviceIFATypeDPID,
					},
				},
			},
		},
		{
			name: `case_insensitive_ifa_type`,
			args: args{
				dvc: &models.DeviceCtx{
					DeviceIFA: `sample_ifa_value`,
					Ext: &models.ExtDevice{
						ExtDevice: openrtb_ext.ExtDevice{
							IFAType: strings.ToUpper(models.DeviceIFATypeDPID),
						},
					},
				},
			},
			want: &models.DeviceCtx{
				DeviceIFA: `sample_ifa_value`,
				IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
				Ext: &models.ExtDevice{
					ExtDevice: openrtb_ext.ExtDevice{
						IFAType: strings.ToUpper(models.DeviceIFATypeDPID),
					},
				},
			},
		},
		{
			name: `ifa_type_present_session_id_present`,
			args: args{
				dvc: &models.DeviceCtx{
					Ext: &models.ExtDevice{
						ExtDevice: openrtb_ext.ExtDevice{
							IFAType: models.DeviceIFATypeDPID,
						},
						SessionID: `sample_session_id`,
					},
				},
			},
			want: &models.DeviceCtx{
				DeviceIFA: `sample_session_id`,
				IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeSESSIONID]),
				Ext: &models.ExtDevice{
					ExtDevice: openrtb_ext.ExtDevice{
						IFAType: models.DeviceIFATypeSESSIONID,
					},
					SessionID: `sample_session_id`,
				},
			},
		},
		{
			name: `ifa_type_present_session_id_missing`,
			args: args{
				dvc: &models.DeviceCtx{
					Ext: &models.ExtDevice{
						ExtDevice: openrtb_ext.ExtDevice{
							IFAType: models.DeviceIFATypeDPID,
						},
					},
				},
			},
			want: &models.DeviceCtx{
				Ext: &models.ExtDevice{
					ExtDevice: openrtb_ext.ExtDevice{},
				},
			},
		},
		{
			name: `ifa_type_missing_session_id_present`,
			args: args{
				dvc: &models.DeviceCtx{
					DeviceIFA: `existing_ifa_id`,
					Ext: &models.ExtDevice{
						ExtDevice: openrtb_ext.ExtDevice{},
						SessionID: `sample_session_id`,
					},
				},
			},
			want: &models.DeviceCtx{
				DeviceIFA: `sample_session_id`,
				IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeSESSIONID]),
				Ext: &models.ExtDevice{
					ExtDevice: openrtb_ext.ExtDevice{
						IFAType: models.DeviceIFATypeSESSIONID,
					},
					SessionID: `sample_session_id`,
				},
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
					Ext: &models.ExtDevice{
						SessionID: `sample_session`,
					},
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
					Ext: &models.ExtDevice{
						SessionID: `sample_session`,
					},
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
		})
	}
}
