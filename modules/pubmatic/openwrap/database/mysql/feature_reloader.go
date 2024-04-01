package mysql

func (db *mySqlDB) GetPublisherFeatureMap() (map[int]int, error) {
	rows, err := db.conn.Query(db.cfg.Queries.GetPublisherFeatureMapQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	publisherFeatureMap := make(map[int]int)
	for rows.Next() {
		var pubid, feature int
		if err := rows.Scan(&pubid, &feature); err == nil {
			publisherFeatureMap[pubid] = feature
		}
	}
	return publisherFeatureMap, nil
}
