package mysql

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/util/ptrutil"
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
		wantErr bool
		setup   func() *sql.DB
	}{
		{
			name:    "empty query in config file",
			want:    nil,
			wantErr: true,
			setup: func() *sql.DB {
				db, _, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
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
			wantErr: false,
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
			wantErr: false,
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
			wantErr: true,
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
			wantErr: false,
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
			wantErr: false,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mySqlDB{
				conn: tt.setup(),
				cfg:  tt.fields.cfg,
			}
			defer db.conn.Close()

			got, err := db.GetAdunitConfig(tt.args.profileID, tt.args.displayVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetAdunitConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
