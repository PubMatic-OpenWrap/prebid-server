package ctv

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/stretchr/testify/assert"
)

func TestUpdateUserExtWithValidValues(t *testing.T) {
	type args struct {
		user *openrtb2.User
	}
	tests := []struct {
		name string
		args args
		want *openrtb2.User
	}{
		{
			name: "test_valid_user_eids",
			args: args{
				user: &openrtb2.User{
					EIDs: []openrtb2.EID{
						{
							Source: "uidapi.com",
							UIDs: []openrtb2.UID{
								{
									ID: "UID2:testUID",
								},
							},
						},
					},
				},
			},
			want: &openrtb2.User{
				EIDs: []openrtb2.EID{
					{
						Source: "uidapi.com",
						UIDs: []openrtb2.UID{
							{
								ID: "testUID",
							},
						},
					},
				},
			},
		},
		{
			name: "test_user_eids_and_user_ext_eids",
			args: args{
				user: &openrtb2.User{
					Ext: json.RawMessage(`{"eids":[{"source":"uidapi.com","uids":[{"id":"UID2:testUID"},{"id":"testUID2"}]},{"source":"euid.eu","uids":[{"id":"testeuid"}]},{"source":"liveramp.com","uids":[{"id":""}]}]}`),
					EIDs: []openrtb2.EID{
						{
							Source: "uidapi.com",
							UIDs: []openrtb2.UID{
								{
									ID: "UID2:testUID",
								},
							},
						},
					},
				},
			},
			want: &openrtb2.User{
				Ext: json.RawMessage(`{"eids":[{"source":"uidapi.com","uids":[{"id":"testUID"},{"id":"testUID2"}]},{"source":"euid.eu","uids":[{"id":"testeuid"}]}]}`),
				EIDs: []openrtb2.EID{
					{
						Source: "uidapi.com",
						UIDs: []openrtb2.UID{
							{
								ID: "testUID",
							},
						},
					},
				},
			},
		},
		{
			name: "test_user_ext_eids",
			args: args{
				user: &openrtb2.User{
					Ext: json.RawMessage(`{"eids":[{"source":"uidapi.com","uids":[{"id":"UID2:testUID"},{"id":"testUID2"}]},{"source":"euid.eu","uids":[{"id":"testeuid"}]},{"source":"liveramp.com","uids":[{"id":""}]}]}`),
				},
			},
			want: &openrtb2.User{
				Ext: json.RawMessage(`{"eids":[{"source":"uidapi.com","uids":[{"id":"testUID"},{"id":"testUID2"}]},{"source":"euid.eu","uids":[{"id":"testeuid"}]}]}`),
			},
		},
		{
			name: "test_user_ext_eids_invalid",
			args: args{
				user: &openrtb2.User{
					Ext: json.RawMessage(`{"eids":[{"source":"uidapi.com","uids":[{"id":"UID2:"},{"id":""}]},{"source":"euid.eu","uids":[{"id":"euid:"}]},{"source":"liveramp.com","uids":[{"id":""}]}]}`),
				},
			},
			want: &openrtb2.User{
				Ext: json.RawMessage(`{}`),
			},
		},
		{
			name: "test_valid_user_eids_invalid",
			args: args{
				user: &openrtb2.User{
					EIDs: []openrtb2.EID{
						{
							Source: "uidapi.com",
							UIDs: []openrtb2.UID{
								{
									ID: "UID2:",
								},
							},
						},
					},
				},
			},
			want: &openrtb2.User{},
		},
		{
			name: "test_valid_user_ext_sessionduration_impdepth",
			args: args{
				user: &openrtb2.User{
					Ext: json.RawMessage(`{"sessionduration":40,"impdepth":10}`),
				},
			},
			want: &openrtb2.User{
				Ext: json.RawMessage(`{"sessionduration":40,"impdepth":10}`),
			},
		},
		{
			name: "test_invalid_user_ext_sessionduration_impdepth",
			args: args{
				user: &openrtb2.User{
					Ext: json.RawMessage(`{
					"sessionduration": 0,
					"impdepth": -10
					}`),
				},
			},
			want: &openrtb2.User{
				Ext: json.RawMessage(`{}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateUserExtWithValidValues(tt.args.user)
			assert.Equal(t, tt.want, tt.args.user)
		})
	}
}
