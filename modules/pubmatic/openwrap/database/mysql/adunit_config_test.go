package mysql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
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
			wantErr: errors.New("context deadline exceeded"),
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
					MaxDbContextTimeout: 1000,
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
					MaxDbContextTimeout: 1000,
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
					MaxDbContextTimeout: 1000,
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
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				profileID:      5890,
				displayVersion: 0,
			},
			want:    nil,
			wantErr: errors.New("unmarshal error adunitconfig"),
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
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				profileID:      5890,
				displayVersion: 1,
			},
			want: &adunitconfig.AdUnitConfig{
				ConfigPattern: "_DIV_",
				Config: map[string]*adunitconfig.AdConfig{
					"default": {
						BidFloor: ptrutil.ToPtr(3.1),
						BidderFilter: &adunitconfig.BidderFilter{
							Filters: []adunitconfig.Filter{
								{
									Bidders:           []string{"A"},
									BiddingConditions: json.RawMessage("\"{ \\\"in\\\": [{ \\\"var\\\": \\\"country\\\"}, [\\\"IND\\\"]]}\""),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"adunitConfig"}).AddRow(`{"configPattern":"_DIV_","config":{"default":{"bidfloor":3.1,"bidderFilter":{"filterConfig":[{"bidders":["A"],"biddingConditions":"{ \"in\": [{ \"var\": \"country\"}, [\"IND\"]]}"}]}}}}`)
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
					MaxDbContextTimeout: 1000,
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
					MaxDbContextTimeout: 1000,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mySqlDB{
				conn: tt.setup(),
				cfg:  tt.fields.cfg,
			}
			defer db.conn.Close()

			got, err := db.GetAdunitConfig(tt.args.profileID, tt.args.displayVersion)
			if tt.wantErr == nil {
				assert.NoError(t, err, tt.name)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error(), tt.name)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
