package mysql

import (
	"context"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func (db *mySqlDB) GetCountryCodesMapping() (models.CountryCodesMapping, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetCountryCodesMapping)
	if err != nil {
		return models.CountryCodesMapping{}, err
	}
	defer rows.Close()

	var (
		countryCode         string
		alpha2Code          string
		alpha3Code          string
		countryCodesMapping = make(models.CountryCodesMapping)
	)
	for rows.Next() {
		if err := rows.Scan(&countryCode, &alpha2Code, &alpha3Code); err != nil {
			glog.Errorf(models.ErrDBRowScanFailed, models.CountryCodesMappingQuery, "", "", err.Error())
			continue
		}
		countryCodesMapping[alpha3Code] = struct {
			Alpha2Code  string `json:"alpha2_code"`
			CountryCode string `json:"country_code"`
		}{
			Alpha2Code:  alpha2Code,
			CountryCode: countryCode,
		}
	}

	if err = rows.Err(); err != nil {
		return models.CountryCodesMapping{}, err
	}
	return countryCodesMapping, nil
}
