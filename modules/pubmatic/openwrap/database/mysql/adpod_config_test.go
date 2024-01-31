package mysql

import (
	"database/sql"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adpodconfig"
)

func TestMySqlDBGetAdpodConfigs(t *testing.T) {
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
		setup   func() *sql.DB
		want    *adpodconfig.AdpodConfig
		wantErr bool
	}{
		{
			name: "Retrieve dynamic adpod configuration from database",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 5,
					Queries: config.Queries{
						GetAdpodConfig:               "^SELECT (.+) FROM wrapper_ad_pod (.+)",
						GetAdpodConfigForLiveVersion: "^SELECT (.+) FROM wrapper_ad_pod (.+) WHERE display_version=0",
					},
				},
			},
			args: args{
				profileID:      123,
				displayVersion: 4,
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pod_type", "s2s_ad_slots_config"}).AddRow("DYNAMIC", `[{"maxduration":60,"maxseq":5,"poddur":180,"minduration":1}]`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_ad_pod (.+)")).WillReturnRows(rows)
				return db
			},
			want: &adpodconfig.AdpodConfig{
				Dynamic: []adpodconfig.Dynamic{
					{
						MaxDuration: 60,
						MinDuration: 1,
						PodDur:      180,
						Maxseq:      5,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Retrieve dynamic adpod configuration from database for live version",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 5,
					Queries: config.Queries{
						GetAdpodConfig:               "^SELECT (.+) FROM wrapper_ad_pod (.+)",
						GetAdpodConfigForLiveVersion: "^SELECT (.+) FROM wrapper_ad_pod (.+) WHERE display_version=0",
					},
				},
			},
			args: args{
				profileID:      123,
				displayVersion: 0,
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pod_type", "s2s_ad_slots_config"}).AddRow("DYNAMIC", `[{"maxduration":60,"maxseq":5,"poddur":180,"minduration":1}]`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_ad_pod (.+) WHERE display_version=0")).WillReturnRows(rows)
				return db
			},
			want: &adpodconfig.AdpodConfig{
				Dynamic: []adpodconfig.Dynamic{
					{
						MaxDuration: 60,
						MinDuration: 1,
						PodDur:      180,
						Maxseq:      5,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Retrieve dynamic adpod configuration from database where rqddurs provided",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 5,
					Queries: config.Queries{
						GetAdpodConfig:               "^SELECT (.+) FROM wrapper_ad_pod (.+)",
						GetAdpodConfigForLiveVersion: "^SELECT (.+) FROM wrapper_ad_pod (.+) WHERE display_version=0",
					},
				},
			},
			args: args{
				profileID:      123,
				displayVersion: 4,
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pod_type", "s2s_ad_slots_config"}).AddRow("DYNAMIC", `[{"maxseq":5,"poddur":600,"rqddurs":[6,60,120,600]}]`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_ad_pod (.+)")).WillReturnRows(rows)
				return db
			},
			want: &adpodconfig.AdpodConfig{
				Dynamic: []adpodconfig.Dynamic{
					{
						PodDur:  600,
						Maxseq:  5,
						Rqddurs: []int{6, 60, 120, 600},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mySqlDB{
				conn: tt.setup(),
				cfg:  tt.fields.cfg,
			}
			got, err := db.GetAdpodConfigs(tt.args.profileID, tt.args.displayVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetAdpodConfigs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mySqlDB.GetAdpodConfigs() = %v, want %v", got, tt.want)
			}
		})
	}
}
