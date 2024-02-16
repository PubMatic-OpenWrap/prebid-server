package mysql

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PubMatic-OpenWrap/prebid-server/util/ptrutil"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/stretchr/testify/assert"
)

func Test_mySqlDB_GetAdunitConfig(t *testing.T) {
	type fields struct {
		cfg config.Database
	}
	type args struct {
		profileID      int
		displayVersion int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *adunitconfig.AdUnitConfig
		wantErr error
		setup   func() *sql.DB
	}{
		{
			name:    "empty query in config file",
			want:    nil,
			wantErr: fmt.Errorf("all expectations were already fulfilled, call to Query '' with args [] was not expected"),
			setup: func() *sql.DB {
				db, _, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				return db
			},
		},
		{
			name: "adunitconfig not found for profile(No rows error)",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAdunitConfigForLiveVersion: "^SELECT (.+) FROM wrapper_media_config (.+) LIVE",
					},
				},
			},
			args: args{
				profileID:      5890,
				displayVersion: 0,
			},
			want:    nil,
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{})
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_media_config (.+) LIVE")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "query with display version id 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAdunitConfigForLiveVersion: "^SELECT (.+) FROM wrapper_media_config (.+) LIVE",
					},
				},
			},
			args: args{
				profileID:      5890,
				displayVersion: 0,
			},
			want: &adunitconfig.AdUnitConfig{
				ConfigPattern: "_AU_",
				Config: map[string]*adunitconfig.AdConfig{
					"default": {BidFloor: ptrutil.ToPtr(2.0)},
				},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"adunitConfig"}).AddRow(`{"config":{"default":{"bidfloor":2}}}`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_media_config (.+) LIVE")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "query with non-zero display version id",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAdunitConfigQuery: "^SELECT (.+) FROM wrapper_media_config (.+)",
					},
				},
			},
			args: args{
				profileID:      5890,
				displayVersion: 1,
			},
			want: &adunitconfig.AdUnitConfig{
				ConfigPattern: "_AU_",
				Config: map[string]*adunitconfig.AdConfig{
					"default": {BidFloor: ptrutil.ToPtr(3.1)},
				},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"adunitConfig"}).AddRow(`{"config":{"default":{"bidfloor":3.1}}}`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_media_config (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "invalid adunitconfig json",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAdunitConfigForLiveVersion: "^SELECT (.+) FROM wrapper_media_config (.+) LIVE",
					},
				},
			},
			args: args{
				profileID:      5890,
				displayVersion: 0,
			},
			want:    nil,
			wantErr: fmt.Errorf("unexpected end of JSON input"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"adunitConfig"}).AddRow(`{`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_media_config (.+) LIVE")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "non-default config pattern in adunitconfig",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAdunitConfigQuery: "^SELECT (.+) FROM wrapper_media_config (.+)",
					},
				},
			},
			args: args{
				profileID:      5890,
				displayVersion: 1,
			},
			want: &adunitconfig.AdUnitConfig{
				ConfigPattern: "_DIV_",
				Config: map[string]*adunitconfig.AdConfig{
					"default": {BidFloor: ptrutil.ToPtr(3.1)},
				},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"adunitConfig"}).AddRow(`{"configPattern": "_DIV_", "config":{"default":{"bidfloor":3.1}}}`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_media_config (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "default adunit not present",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAdunitConfigQuery: "^SELECT (.+) FROM wrapper_media_config (.+)",
					},
				},
			},
			args: args{
				profileID:      5890,
				displayVersion: 1,
			},
			want: &adunitconfig.AdUnitConfig{
				ConfigPattern: "_DIV_",
				Config: map[string]*adunitconfig.AdConfig{
					"default": {},
					"abc":     {BidFloor: ptrutil.ToPtr(3.1)},
				},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"adunitConfig"}).AddRow(`{"configPattern": "_DIV_", "config":{"abc":{"bidfloor":3.1}}}`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_media_config (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "config key not present",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAdunitConfigQuery: "^SELECT (.+) FROM wrapper_media_config (.+)",
					},
				},
			},
			args: args{
				profileID:      5890,
				displayVersion: 1,
			},
			want: &adunitconfig.AdUnitConfig{
				ConfigPattern: "_DIV_",
				Config: map[string]*adunitconfig.AdConfig{
					"default": {},
				},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"adunitConfig"}).AddRow(`{"configPattern": "_DIV_"}`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_media_config (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "adunit config unmarshal check",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAdunitConfigQuery: "^SELECT (.+) FROM wrapper_media_config (.+)",
					},
				},
			},
			args: args{
				profileID:      5890,
				displayVersion: 1,
			},
			want:    nil,
			wantErr: fmt.Errorf("json: cannot unmarshal string into Go struct field Banner.config.banner.enabled of type bool"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"adunitConfig"}).AddRow(`{"configPattern":"_AU_","config":{"default":{"banner":{"enabled":false},"video":{"enabled":false}},"/15671365/MG_VideoAdUnit":{"native":{"enabled":true},"banner":{"config":{"clientconfig":{"refreshinterval":30}},"enabled":"true"},"video":{"config":{"battr":[6,7],"skipafter":15,"maxduration":53,"context":"instream","playerSize":[640,480],"skip":1,"connectiontype":[1,2,6],"skipmin":10,"minduration":11,"mimes":["video/mp4","video/x-flv"]},"enabled":true}}}}`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_media_config (.+)")).WillReturnRows(rows)
				return db
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mySqlDB{
				conn: tt.setup(),
				cfg:  tt.fields.cfg,
			}
			defer db.conn.Close()

			got, err := db.GetAdunitConfig(tt.args.profileID, tt.args.displayVersion)
			if err != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
