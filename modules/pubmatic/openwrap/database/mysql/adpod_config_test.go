package mysql

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adpodconfig"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
	"github.com/stretchr/testify/assert"
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
		wantErr error
	}{
		{
			name: "Retrieve dynamic adpod configuration from database",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 5,
					Queries: config.Queries{
						GetAdpodConfig:           "^SELECT (.+) FROM ad_pod (.+)",
						DisplayVersionInnerQuery: "^SELECT (.+) FROM version (.+)",
					},
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
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM version (.+)")).WithArgs(123, 4, 5890).WillReturnRows(rowsWrapperVersion)
				rows := sqlmock.NewRows([]string{"pod_type", "s2s_ad_slots_config"}).AddRow("DYNAMIC", `[{"maxduration":60,"maxseq":5,"poddur":180,"minduration":1}]`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM ad_pod (.+)")).WithArgs(4444).WillReturnRows(rows)
				return db
			},
			want: &adpodconfig.AdpodConfig{
				Dynamic: []adpodconfig.Dynamic{
					{
						MaxDuration: 60,
						MinDuration: 1,
						PodDur:      180,
						MaxSeq:      5,
					},
				},
			},
			wantErr: nil,
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
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM version (.+) LIVE")).WithArgs(123, 5890).WillReturnRows(rowsWrapperVersion)
				rows := sqlmock.NewRows([]string{"pod_type", "s2s_ad_slots_config"}).AddRow("DYNAMIC", `[{"maxduration":60,"maxseq":5,"poddur":180,"minduration":1}]`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM ad_pod (.+)")).WithArgs(4444).WillReturnRows(rows)
				return db
			},
			want: &adpodconfig.AdpodConfig{
				Dynamic: []adpodconfig.Dynamic{
					{
						MaxDuration: 60,
						MinDuration: 1,
						PodDur:      180,
						MaxSeq:      5,
					},
				},
			},
			wantErr: nil,
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
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM version (.+)")).WithArgs(123, 4, 5890).WillReturnRows(rowsWrapperVersion)
				rows := sqlmock.NewRows([]string{"pod_type", "s2s_ad_slots_config"}).AddRow("DYNAMIC", `[{"maxseq":5,"poddur":600,"rqddurs":[6,60,120,600]}]`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM ad_pod (.+)")).WithArgs(4444).WillReturnRows(rows)
				return db
			},
			want: &adpodconfig.AdpodConfig{
				Dynamic: []adpodconfig.Dynamic{
					{
						PodDur:  600,
						MaxSeq:  5,
						RqdDurs: []int64{6, 60, 120, 600},
					},
				},
			},
			wantErr: nil,
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
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM version (.+)")).WithArgs(123, 4, 5890).WillReturnRows(rowsWrapperVersion)
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
						MaxSeq:      5,
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
						MaxSeq:      ptrutil.ToPtr(int64(3)),
						PodDur:      ptrutil.ToPtr(int64(60)),
					},
				},
			},
			wantErr: nil,
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
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM version (.+)")).WithArgs(123, 4, 5890).WillReturnRows(rowsWrapperVersion)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM ad_pod (.+)")).WillReturnError(errors.New("context deadline exceeded"))
				return db
			},
			want:    nil,
			wantErr: errors.New("GetAdpodConfigQuery Failure Error: context deadline exceeded"),
		},
		{
			name: "Error in row scan",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 5,
					Queries: config.Queries{
						GetAdpodConfig:           "^SELECT (.+) FROM ad_pod (.+)",
						DisplayVersionInnerQuery: "^SELECT (.+) FROM version (.+)",
					},
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
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM version (.+)")).WithArgs(123, 4, 5890).WillReturnRows(rowsWrapperVersion)
				rows := sqlmock.NewRows([]string{"pod_type", "s2s_ad_slots_config"}).
					AddRow("DYNAMIC", `[{"maxduration":60,"maxseq":5,"poddur":180,"minduration":1}]`).
					AddRow("HYBRID", `[{"maxduration":20,"minduration":5},{"maxduration":20,"maxseq":3,"poddur":60,"minduration":5}]`).
					AddRow("STRUCTURED", `[{"maxduration":20,"minduration":5}]`)
				rows = rows.RowError(1, errors.New("error in row scan"))
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM ad_pod (.+)")).WillReturnRows(rows)
				return db
			},
			want:    nil,
			wantErr: errors.New("GetAdpodConfigQuery Failure Error: error in row scan"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mySqlDB{
				conn: tt.setup(),
				cfg:  tt.fields.cfg,
			}
			got, err := db.GetAdpodConfig(tt.args.pubId, tt.args.profileID, tt.args.displayVersion)
			if tt.wantErr == nil {
				assert.NoError(t, err, tt.name)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error(), tt.name)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mySqlDB.GetAdpodConfigs() = %v, want %v", got, tt.want)
			}
		})
	}
}
