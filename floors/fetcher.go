package floors

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	validator "github.com/asaskevich/govalidator"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
	"golang.org/x/net/context/ctxhttp"
)

// fetchResult defines the contract for fetched floors results
type fetchedFloors struct {
	FetchedJSON []fetchResult `json:"fetchresult,omitempty"`
}

// fetchResult defines the contract for fetched floors results
type fetchResult struct {
	PriceFloors openrtb_ext.PriceFloorRules `json:"pricefloors,omitempty"`
	FetchStatus int                         `json:"fetchstatus,omitempty"`
}

var fetchInProgress map[string]bool

func fetchInit() {
	fetchInProgress = make(map[string]bool)
}

func getBoolPtr(val bool) *bool {
	return &val
}

// fetchAccountFloors this function fetch floors JSON for given account
var fetchAccountFloors = func(account config.Account) *fetchResult {

	//	var fetchedResults fetchResult
	// Check for Rules in cache

	// fetch floors JSON
	return fetchPriceFloorRules(account)
}

func fetchPriceFloorRules(account config.Account) *fetchResult {
	// If fetch is disabled
	fetchConfig := account.PriceFloors.Fetch
	if !fetchConfig.Enabled {
		return &fetchResult{
			FetchStatus: openrtb_ext.FetchNone,
		}
	}

	if !validator.IsURL(fetchConfig.URL) {
		return &fetchResult{
			FetchStatus: openrtb_ext.FetchError,
		}
	}

	_, fetchInprogress := fetchInProgress[fetchConfig.URL]
	if !fetchInprogress {
		fetchPriceFloorRulesAsynchronous(account)
	}

	// Rules not present in cache, fetch rules asynchronously
	return &fetchResult{
		FetchStatus: openrtb_ext.FetchInprogress,
	}
}

func fetchPriceFloorRulesAsynchronous(account config.Account) []error {

	var errList []error
	start := time.Now()

	ctx := context.Background()

	timeout := (time.Duration(account.PriceFloors.Fetch.Timeout) * time.Millisecond)
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, start.Add(timeout))
		defer cancel()
	}

	httpReq, err := http.NewRequest("GET", account.PriceFloors.Fetch.URL, nil)
	if err != nil {
		return []error{err}
	}

	httpResp, err := ctxhttp.Do(ctx, &http.Client{}, httpReq)
	if err != nil {
		return []error{err}
	}

	if httpResp.StatusCode == 200 {
		respBody, err := ioutil.ReadAll(httpResp.Body)
		if err != nil {
			return []error{err}
		}
		defer httpResp.Body.Close()

		if len(respBody) > account.PriceFloors.Fetch.MaxFileSize {
			return []error{fmt.Errorf("Receieved more than MaxFileSize")}
		}

		var priceFloors openrtb_ext.PriceFloorRules

		err = json.Unmarshal(respBody, &priceFloors)
		if err != nil {
			return []error{fmt.Errorf("Error in JSON Unmarshall ")}
		}

		errList = validatePriceFloorRules(priceFloors, account.PriceFloors.Fetch)

		if priceFloors.Data != nil && len(priceFloors.Data.ModelGroups) > 0 {
			// Push floors JSON to cache

			// Create periodic fetching JOB
		}
	}
	return errList
}

func validatePriceFloorRules(priceFloors openrtb_ext.PriceFloorRules, fetchConfig config.AccountFloorFetch) []error {
	floorData := priceFloors.Data

	var err []error
	if floorData == nil {
		return []error{fmt.Errorf("Empty data in floors JSON  in JSON Unmarshall ")}
	}

	var validModelGroups []openrtb_ext.PriceFloorModelGroup
	for _, modelGroup := range floorData.ModelGroups {
		if len(modelGroup.Values) > fetchConfig.MaxRules {
			err = append(err, fmt.Errorf("Number of rules = %v in modelgroup are greater than limit = %v", len(modelGroup.Values), fetchConfig.MaxRules))
		}
		validModelGroups = append(validModelGroups, modelGroup)
	}
	floorData.ModelGroups = validModelGroups
	return err
}
