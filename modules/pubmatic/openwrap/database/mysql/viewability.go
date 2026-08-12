package mysql

import (
	"context"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func (db *mySqlDB) GetInViewEnabledPublishers() (map[int]struct{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetInViewEnabledPublishersQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	inViewEnabledPublishers := make(map[int]struct{})
	for rows.Next() {
		var pubID int
		if err := rows.Scan(&pubID); err != nil {
			glog.Errorf(models.ErrDBRowScanFailed, models.InViewEnabledPublishersQuery, "", "", err.Error())
			continue
		}
		inViewEnabledPublishers[pubID] = struct{}{}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return inViewEnabledPublishers, nil
}
