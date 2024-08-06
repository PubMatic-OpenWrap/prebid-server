package mysql

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adpodconfig"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
)

func TestMySqlDBGetAdpodConfigs(t *testing.T) {
	type fields struct {
		cfg config.Database
	}
	type args struct {
		pubId          int
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
						GetAdpodConfig:           "^SELECT (.+) FROM  ad_pod (.+)",
						DisplayVersionInnerQuery: "^SELECT (.+) FROM  version (.+)",
					},
					MaxQueryExecutionTimeout: 500,
				},
			},
			args: args{
				pubId:          5890,
				profileID:      123,
				displayVersion: 4,
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId", "platform", "type"}).
					AddRow("4444", "4", "ctv", "1")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM version (.+)")).WithArgs(500, 123, 4, 5890).WillReturnRows(rowsWrapperVersion)

				rows := sqlmock.NewRows([]string{"pod_type", "s2s_ad_slots_config"}).AddRow("DYNAMIC", `[{"maxduration":60,"maxseq":5,"poddur":180,"minduration":1}]`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM ad_pod (.+)")).WithArgs(500, 4444).WillReturnRows(rows)
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
						GetAdpodConfig:        "^SELECT (.+) FROM ad_pod (.+)",
						LiveVersionInnerQuery: "^SELECT (.+) FROM version (.+) LIVE",
					},
					MaxQueryExecutionTimeout: 500,
				},
			},
			args: args{
				pubId:          5890,
				profileID:      123,
				displayVersion: 0,
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId", "platform", "type"}).AddRow("4444", "4", "ctv", "1")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM version (.+) LIVE")).WithArgs(500, 123, 5890).WillReturnRows(rowsWrapperVersion)

				rows := sqlmock.NewRows([]string{"pod_type", "s2s_ad_slots_config"}).AddRow("DYNAMIC", `[{"maxduration":60,"maxseq":5,"poddur":180,"minduration":1}]`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM ad_pod (.+)")).WithArgs(500, 4444).WillReturnRows(rows)
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
						GetAdpodConfig:           "^SELECT (.+) FROM ad_pod (.+)",
						DisplayVersionInnerQuery: "^SELECT (.+) FROM version (.+)",
					},
					MaxQueryExecutionTimeout: 500,
				},
			},
			args: args{
				pubId:          5890,
				profileID:      123,
				displayVersion: 4,
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId", "platform", "type"}).AddRow("4444", "4", "ctv", "1")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM version (.+)")).WithArgs(500, 123, 4, 5890).WillReturnRows(rowsWrapperVersion)

				rows := sqlmock.NewRows([]string{"pod_type", "s2s_ad_slots_config"}).AddRow("DYNAMIC", `[{"maxseq":5,"poddur":600,"rqddurs":[6,60,120,600]}]`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM ad_pod (.+)")).WithArgs(500, 4444).WillReturnRows(rows)
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
		{
			name: "Retrieve dynamic adpod configuration from database for all types",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 5,
					Queries: config.Queries{
						GetAdpodConfig:           "^SELECT (.+) FROM ad_pod (.+)",
						DisplayVersionInnerQuery: "^SELECT (.+) FROM version (.+)",
					},
					MaxQueryExecutionTimeout: 500,
				},
			},
			args: args{
				pubId:          5890,
				profileID:      123,
				displayVersion: 4,
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId", "platform", "type"}).AddRow("4444", "4", "ctv", "1")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM version (.+)")).WithArgs(500, 123, 4, 5890).WillReturnRows(rowsWrapperVersion)

				rows := sqlmock.NewRows([]string{"pod_type", "s2s_ad_slots_config"}).
					AddRow("DYNAMIC", `[{"maxduration":60,"maxseq":5,"poddur":180,"minduration":1}]`).
					AddRow("HYBRID", `[{"maxduration":20,"minduration":5},{"maxduration":20,"maxseq":3,"poddur":60,"minduration":5}]`).
					AddRow("STRUCTURED", `[{"maxduration":20,"minduration":5}]`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM ad_pod (.+)")).WillReturnRows(rows)
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
				Structured: []adpodconfig.Structured{
					{
						MinDuration: 5,
						MaxDuration: 20,
					},
				},
				Hybrid: []adpodconfig.Hybrid{
					{
						MinDuration: 5,
						MaxDuration: 20,
					},
					{
						MaxDuration: 20,
						MinDuration: 5,
						Maxseq:      ptrutil.ToPtr(3),
						PodDur:      ptrutil.ToPtr(60),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "No adpod configuration in database",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 50,
					Queries: config.Queries{
						GetAdpodConfig:           "^SELECT (.+) FROM ad_pod (.+)",
						DisplayVersionInnerQuery: "^SELECT (.+) FROM version (.+)",
					},
					MaxQueryExecutionTimeout: 500,
				},
			},
			args: args{
				pubId:          5890,
				profileID:      123,
				displayVersion: 4,
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId", "platform", "type"}).AddRow("4444", "4", "ctv", "1")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM version (.+)")).WithArgs(500, 123, 4, 5890).WillReturnRows(rowsWrapperVersion)

				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM ad_pod (.+)")).WillReturnError(errors.New("context deadline exceeded"))
				return db
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mySqlDB{
				conn: tt.setup(),
				cfg:  tt.fields.cfg,
			}
			got, err := db.GetAdpodConfig(tt.args.pubId, tt.args.profileID, tt.args.displayVersion)
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
