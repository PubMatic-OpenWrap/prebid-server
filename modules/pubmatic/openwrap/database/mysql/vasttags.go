package mysql

import (
	"fmt"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// GetPublisherVASTTags - Method to get vast tags associated with publisher id from giym DB
func (db *mySqlDB) GetPublisherVASTTags(pubID int) (models.PublisherVASTTags, error) {

	/*
		//TOOD:VIRAL Remove Hook once UI/API changes are in place
		if out := vastTagHookPublisherVASTTags(rtbReqId, pubID); nil != out {
			return out, nil
		}
	*/

	getActiveVASTTagsQuery := fmt.Sprintf(db.cfg.Queries.GetPublisherVASTTagsQuery, pubID)

	rows, err := db.conn.Query(getActiveVASTTagsQuery)
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
	return vasttags, nil
}
