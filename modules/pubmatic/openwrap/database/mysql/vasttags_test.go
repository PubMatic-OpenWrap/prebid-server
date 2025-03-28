package mysql

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func Test_mySqlDB_GetPublisherVASTTags(t *testing.T) {
	type fields struct {
		cfg config.Database
	}
	type args struct {
		pubID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.PublisherVASTTags
		wantErr error
		setup   func() *sql.DB
	}{
		// {
		// 	name:    "empty query in config file",
		// 	want:    map[int]*models.VASTTag{},
		// 	wantErr: errors.New("context deadline exceeded"),
		// 	setup: func() *sql.DB {
		// 		db, _, err := sqlmock.New()
		// 		if err != nil {
		// 			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		// 		}
		// 		return db
		// 	},
		// },
		// {
		// 	name: "invalid vast tag partnerId",
		// 	fields: fields{
		// 		cfg: config.Database{
		// 			Queries: config.Queries{
		// 				GetPublisherVASTTagsQuery: "^SELECT (.+) FROM wrapper_publisher_partner_vast_tag (.+)",
		// 			},
		// 			MaxDbContextTimeout: 1000,
		// 		},
		// 	},
		// 	args: args{
		// 		pubID: 5890,
		// 	},
		// 	want: models.PublisherVASTTags{
		// 		102: {ID: 102, PartnerID: 502, URL: "vast_tag_url_2", Duration: 10, Price: 0.0},
		// 		103: {ID: 103, PartnerID: 501, URL: "vast_tag_url_1", Duration: 30, Price: 3.0},
		// 	},
		// 	wantErr: nil,
		// 	setup: func() *sql.DB {
		// 		db, mock, err := sqlmock.New()
		// 		if err != nil {
		// 			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		// 		}
		// 		rows := sqlmock.NewRows([]string{"id", "partnerId", "url", "duration", "price"}).
		// 			AddRow(101, "501_12", "vast_tag_url_1", 15, 2.0).
		// 			AddRow(102, 502, "vast_tag_url_2", 10, 0.0).
		// 			AddRow(103, 501, "vast_tag_url_1", 30, 3.0)
		// 		mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_publisher_partner_vast_tag (.+)")).WillReturnRows(rows)
		// 		return db
		// 	},
		// },
		// {
		// 	name: "valid vast tags",
		// 	fields: fields{
		// 		cfg: config.Database{
		// 			Queries: config.Queries{
		// 				GetPublisherVASTTagsQuery: "^SELECT (.+) FROM wrapper_publisher_partner_vast_tag (.+)",
		// 			},
		// 			MaxDbContextTimeout: 1000,
		// 		},
		// 	},
		// 	args: args{
		// 		pubID: 5890,
		// 	},
		// 	want: models.PublisherVASTTags{
		// 		101: {ID: 101, PartnerID: 501, URL: "vast_tag_url_1", Duration: 15, Price: 2.0},
		// 		102: {ID: 102, PartnerID: 502, URL: "vast_tag_url_2", Duration: 10, Price: 0.0},
		// 		103: {ID: 103, PartnerID: 501, URL: "vast_tag_url_1", Duration: 30, Price: 3.0},
		// 	},
		// 	wantErr: nil,
		// 	setup: func() *sql.DB {
		// 		db, mock, err := sqlmock.New()
		// 		if err != nil {
		// 			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		// 		}
		// 		rows := sqlmock.NewRows([]string{"id", "partnerId", "url", "duration", "price"}).
		// 			AddRow(101, 501, "vast_tag_url_1", 15, 2.0).
		// 			AddRow(102, 502, "vast_tag_url_2", 10, 0.0).
		// 			AddRow(103, 501, "vast_tag_url_1", 30, 3.0)
		// 		mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_publisher_partner_vast_tag (.+)")).WillReturnRows(rows)
		// 		return db
		// 	},
		// },
		{
			name: "error in row scan",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPublisherVASTTagsQuery: "^SELECT (.+) FROM wrapper_publisher_partner_vast_tag (.+)",
					},
					MaxDbContextTimeout: 100000,
				},
			},
			args: args{
				pubID: 5890,
			},
			want:    models.PublisherVASTTags{101: &models.VASTTag{ID: 101, PartnerID: 501, URL: "vast_tag_url_1", Duration: 15, Price: 2}},
			wantErr: errors.New("error in row scan"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"id", "partnerId", "url", "duration", "price"}).
					AddRow(101, 501, "vast_tag_url_1", 15, 2.0).
					AddRow(102, 502, "vast_tag_url_2", 10, 0.0).
					AddRow(103, 501, "vast_tag_url_1", 30, 3.0)
				rows = rows.RowError(1, errors.New("error in row scan"))
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_publisher_partner_vast_tag (.+)")).WillReturnRows(rows)
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
			got, err := db.GetPublisherVASTTags(tt.args.pubID)
			if tt.wantErr == nil {
				assert.NoError(t, err, tt.name)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error(), tt.name)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
