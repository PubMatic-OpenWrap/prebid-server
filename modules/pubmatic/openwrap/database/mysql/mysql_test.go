package mysql

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		conn *sql.DB
		cfg  config.Database
	}
	tests := []struct {
		name  string
		args  args
		setup func() *sql.DB
	}{
		{
			name: "test",
			args: args{
				cfg: config.Database{},
			},
			setup: func() *sql.DB {
				db, _, _ := sqlmock.New()
				return db
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.conn = tt.setup()
			db := New(tt.args.conn, tt.args.cfg)
			assert.NotNil(t, db, "db should not be nil")
		})
	}
}
