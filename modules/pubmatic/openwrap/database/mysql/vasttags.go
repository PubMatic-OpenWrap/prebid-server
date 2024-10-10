package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

// GetPublisherVASTTags - Method to get vast tags associated with publisher id from giym DB
func (db *mySqlDB) GetPublisherVASTTags(pubID int) (models.PublisherVASTTags, error) {
	getActiveVASTTagsQuery := fmt.Sprintf(db.cfg.Queries.GetPublisherVASTTagsQuery, pubID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, getActiveVASTTagsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vasttags := models.PublisherVASTTags{}
	for rows.Next() {
		var vastTag models.VASTTag
		if err := rows.Scan(&vastTag.ID, &vastTag.PartnerID, &vastTag.URL, &vastTag.Duration, &vastTag.Price); err == nil {
			vasttags[vastTag.ID] = &vastTag
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return vasttags, nil
}
