package mysql

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
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
			name: "invalid vast tag partnerId",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPublisherVASTTagsQuery: `SELECT wvt.id AS id, wvt.partner_id AS partnerId, wvt.url AS url, wvt.duration AS duration, IFNULL(wvt.price,0.0) AS price FROM wrapper_publisher_partner_vast_tag wvt WHERE wvt.pub_id = %d AND wvt.deleted = 0 AND wvt.status="live" ORDER BY wvt.partner_id`,
					},
				},
			},
			args: args{
				pubID: 5890,
			},
			want: models.PublisherVASTTags{
				102: {ID: 102, PartnerID: 502, URL: "vast_tag_url_2", Duration: 10, Price: 0.0},
				103: {ID: 103, PartnerID: 501, URL: "vast_tag_url_1", Duration: 30, Price: 3.0},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"id", "partnerId", "url", "duration", "price"}).
					AddRow(101, "501_12", "vast_tag_url_1", 15, 2.0).
					AddRow(102, 502, "vast_tag_url_2", 10, 0.0).
					AddRow(103, 501, "vast_tag_url_1", 30, 3.0)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT wvt.id AS id, wvt.partner_id AS partnerId, wvt.url AS url, wvt.duration AS duration, IFNULL(wvt.price,0.0) AS price FROM wrapper_publisher_partner_vast_tag wvt WHERE wvt.pub_id = 5890 AND wvt.deleted = 0 AND wvt.status="live" ORDER BY wvt.partner_id`)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "valid vast tags",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPublisherVASTTagsQuery: `SELECT wvt.id AS id, wvt.partner_id AS partnerId, wvt.url AS url, wvt.duration AS duration, IFNULL(wvt.price,0.0) AS price FROM wrapper_publisher_partner_vast_tag wvt WHERE wvt.pub_id = %d AND wvt.deleted = 0 AND wvt.status="live" ORDER BY wvt.partner_id`,
					},
				},
			},
			args: args{
				pubID: 5890,
			},
			want: models.PublisherVASTTags{
				101: {ID: 101, PartnerID: 501, URL: "vast_tag_url_1", Duration: 15, Price: 2.0},
				102: {ID: 102, PartnerID: 502, URL: "vast_tag_url_2", Duration: 10, Price: 0.0},
				103: {ID: 103, PartnerID: 501, URL: "vast_tag_url_1", Duration: 30, Price: 3.0},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"id", "partnerId", "url", "duration", "price"}).
					AddRow(101, 501, "vast_tag_url_1", 15, 2.0).
					AddRow(102, 502, "vast_tag_url_2", 10, 0.0).
					AddRow(103, 501, "vast_tag_url_1", 30, 3.0)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT wvt.id AS id, wvt.partner_id AS partnerId, wvt.url AS url, wvt.duration AS duration, IFNULL(wvt.price,0.0) AS price FROM wrapper_publisher_partner_vast_tag wvt WHERE wvt.pub_id = 5890 AND wvt.deleted = 0 AND wvt.status="live" ORDER BY wvt.partner_id`)).WillReturnRows(rows)
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
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetPublisherVASTTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
