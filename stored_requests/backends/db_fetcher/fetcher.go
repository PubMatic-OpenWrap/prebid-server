package db_fetcher

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/PubMatic-OpenWrap/prebid-server/stored_requests"
	"github.com/golang/glog"
)

func NewFetcher(db *sql.DB, queryMaker func(int) (string, error)) stored_requests.Fetcher {
	return &dbFetcher{
		db:         db,
		queryMaker: queryMaker,
	}
}

// dbFetcher fetches Stored Requests from a database. This should be instantiated through the NewFetcher() function.
type dbFetcher struct {
	db         *sql.DB
	queryMaker func(int) (string, error)
}

func (fetcher *dbFetcher) FetchRequests(ctx context.Context, ids []string) (map[string]json.RawMessage, []error) {
	if len(ids) < 1 {
		return nil, nil
	}

	query, err := fetcher.queryMaker(len(ids))
	if err != nil {
		return nil, []error{err}
	}

	idInterfaces := make([]interface{}, len(ids))
	for i := 0; i < len(ids); i++ {
		idInterfaces[i] = ids[i]
	}

	rows, err := fetcher.db.QueryContext(ctx, query, idInterfaces...)
	if err != nil {
		ctxErr := ctx.Err()
		// This query might fail if the user chose an extremely short timeout.
		// We don't care about these... but there may also be legit connection issues.
		// Log any other errors so we have some idea what's going on.
		if ctxErr == nil || ctxErr != context.DeadlineExceeded {
			glog.Errorf("Error reading from Stored Request DB: %s", err.Error())
		}
		return nil, []error{err}
	}
	defer rows.Close()

	reqData := make(map[string]json.RawMessage, len(ids))
	var errs []error = nil
	for rows.Next() {
		var id string
		var thisReqData []byte
		if err := rows.Scan(&id, &thisReqData); err != nil {
			errs = append(errs, err)
		}

		reqData[id] = thisReqData
	}
	if rows.Err() != nil {
		errs = append(errs, rows.Err())
	}

	for _, id := range ids {
		if _, ok := reqData[id]; !ok {
			errs = append(errs, fmt.Errorf(`Stored Request with ID="%s" not found.`, id))
		}
	}

	return reqData, errs
}
