package models

import (
	"testing"

	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestGetSlotName(t *testing.T) {
	type args struct {
		tagId  string
		impExt *ImpExtension
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Slot_name_from_gpid",
			args: args{
				tagId: "some-tagid",
				impExt: &ImpExtension{
					GpId: "some-gpid",
				},
			},
			want: "some-gpid",
		},
		{
			name: "Slot_name_from_tagid",
			args: args{
				tagId: "some-tagid",
				impExt: &ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "some-pbadslot",
					},
				},
			},
			want: "some-tagid",
		},
		{
			name: "Slot_name_from_pbadslot",
			args: args{
				tagId: "",
				impExt: &ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "some-pbadslot",
					},
				},
			},
			want: "some-pbadslot",
		},
		{
			name: "Slot_name_from_stored_request_id",
			args: args{
				tagId: "",
				impExt: &ImpExtension{
					Prebid: openrtb_ext.ExtImpPrebid{
						StoredRequest: &openrtb_ext.ExtStoredRequest{
							ID: "stored-req-id",
						},
					},
				},
			},
			want: "stored-req-id",
		},
		{
			name: "imp_ext_nil_slot_name_from_tag_id",
			args: args{
				tagId:  "some-tagid",
				impExt: nil,
			},
			want: "some-tagid",
		},
		{
			name: "empty_slot_name",
			args: args{
				tagId:  "",
				impExt: &ImpExtension{},
			},
			want: "",
		},
		{
			name: "all_level_information_is_present_slot_name_picked_by_preference",
			args: args{
				tagId: "some-tagid",
				impExt: &ImpExtension{
					GpId: "some-gpid",
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "some-pbadslot",
					},
					Prebid: openrtb_ext.ExtImpPrebid{
						StoredRequest: &openrtb_ext.ExtStoredRequest{
							ID: "stored-req-id",
						},
					},
				},
			},
			want: "some-gpid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSlotName(tt.args.tagId, tt.args.impExt)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestGetAdunitName(t *testing.T) {
	type args struct {
		tagId  string
		impExt *ImpExtension
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "adunit_from_adserver_slot",
			args: args{
				tagId: "some-tagid",
				impExt: &ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "some-pbadslot",
						AdServer: &openrtb_ext.ExtImpDataAdServer{
							Name:   GamAdServer,
							AdSlot: "gam-unit",
						},
					},
				},
			},
			want: "gam-unit",
		},
		{
			name: "adunit_from_pbadslot",
			args: args{
				tagId: "some-tagid",
				impExt: &ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "some-pbadslot",
						AdServer: &openrtb_ext.ExtImpDataAdServer{
							Name:   GamAdServer,
							AdSlot: "",
						},
					},
				},
			},
			want: "some-pbadslot",
		},
		{
			name: "adunit_from_pbadslot_when_gam_is_absent",
			args: args{
				tagId: "some-tagid",
				impExt: &ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "some-pbadslot",
						AdServer: &openrtb_ext.ExtImpDataAdServer{
							Name:   "freewheel",
							AdSlot: "freewheel-unit",
						},
					},
				},
			},
			want: "some-pbadslot",
		},
		{
			name: "adunit_from_TagId",
			args: args{
				tagId: "some-tagid",
				impExt: &ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "",
						AdServer: &openrtb_ext.ExtImpDataAdServer{
							Name:   GamAdServer,
							AdSlot: "",
						},
					},
				},
			},
			want: "some-tagid",
		},
		{
			name: "adunit_from_TagId_imp_ext_nil",
			args: args{
				tagId:  "some-tagid",
				impExt: nil,
			},
			want: "some-tagid",
		},
		{
			name: "adunit_from_TagId_imp_ext_nil",
			args: args{
				tagId:  "some-tagid",
				impExt: &ImpExtension{},
			},
			want: "some-tagid",
		},
		{
			name: "all_level_information_is_present_adunit_name_picked_by_preference",
			args: args{
				tagId: "some-tagid",
				impExt: &ImpExtension{
					GpId: "some-gpid",
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "some-pbadslot",
						AdServer: &openrtb_ext.ExtImpDataAdServer{
							Name:   GamAdServer,
							AdSlot: "gam-unit",
						},
					},
				},
			},
			want: "gam-unit",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAdunitName(tt.args.tagId, tt.args.impExt)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
