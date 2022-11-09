package floors

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	validator "github.com/asaskevich/govalidator"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
	"golang.org/x/net/context/ctxhttp"
)

// fetchReult defines the contract for fetched floors results
type fetchReult struct {
	priceFloors openrtb_ext.PriceFloorRules `json:"pricefloors,omitempty"`
	fetchStatus int                         `json:"fetchstatus,omitempty"`
}

var fetchInProgress map[string]bool

func fetchInit() {
	fetchInProgress = make(map[string]bool)
}

func getBoolPtr(val bool) *bool {
	return &val
}

// fetchAccountFloors this function fetch floors JSON for given account
var fetchAccountFloors = func(account config.Account) *fetchReult {

	maxFetchResults := 3
	sampleFloors := [3]fetchReult{
		{
			fetchStatus: openrtb_ext.FetchSuccess,
			priceFloors: openrtb_ext.PriceFloorRules{
				FloorMin:           3,
				FloorMinCur:        "INR",
				Enabled:            getBoolPtr(true),
				PriceFloorLocation: openrtb_ext.FetchLocation,
				Enforcement: &openrtb_ext.PriceFloorEnforcement{
					EnforcePBS:  getBoolPtr(true),
					EnforceRate: 100,
					FloorDeals:  getBoolPtr(true),
				},
				Data: &openrtb_ext.PriceFloorData{
					Currency: "USD",
					ModelGroups: []openrtb_ext.PriceFloorModelGroup{
						{
							ModelVersion: "1 dynamic model 1",
							Currency:     "USD",
							Values: map[string]float64{
								"banner|300x600|www.website5.com": 5,
								"*|*|*":                           7,
							},
							Schema: openrtb_ext.PriceFloorSchema{
								Fields:    []string{"mediaType", "size", "domain"},
								Delimiter: "|",
							},
						},
					},
				},
			},
		},
		{
			fetchStatus: openrtb_ext.FetchSuccess,
			priceFloors: openrtb_ext.PriceFloorRules{
				FloorMin:           5,
				FloorMinCur:        "EUR",
				Enabled:            getBoolPtr(true),
				PriceFloorLocation: openrtb_ext.FetchLocation,
				Enforcement: &openrtb_ext.PriceFloorEnforcement{
					EnforcePBS:  getBoolPtr(true),
					EnforceRate: 100,
					FloorDeals:  getBoolPtr(true),
				},
				Data: &openrtb_ext.PriceFloorData{
					Currency: "USD",
					ModelGroups: []openrtb_ext.PriceFloorModelGroup{
						{
							ModelVersion: "2 dynamic model 1",
							Currency:     "USD",
							Values: map[string]float64{
								"banner|300x600|*": 5,
								"*|*|*":            7,
							},
							Schema: openrtb_ext.PriceFloorSchema{
								Fields:    []string{"mediaType", "size", "domain"},
								Delimiter: "|",
							},
						},
						{
							ModelVersion: "2 dynamic model 2",
							Currency:     "USD",
							Values: map[string]float64{
								"banner|300x250|www.website.com": 15,
								"*|*|*":                          17,
							},
							Schema: openrtb_ext.PriceFloorSchema{
								Fields:    []string{"mediaType", "size", "domain"},
								Delimiter: "|",
							},
						},
					},
				},
			},
		},
		{
			fetchStatus: openrtb_ext.FetchSuccess,
			priceFloors: openrtb_ext.PriceFloorRules{
				FloorMin:           7,
				FloorMinCur:        "USD",
				Enabled:            getBoolPtr(false),
				PriceFloorLocation: openrtb_ext.FetchLocation,
				Enforcement: &openrtb_ext.PriceFloorEnforcement{
					EnforcePBS:  getBoolPtr(true),
					EnforceRate: 100,
					FloorDeals:  getBoolPtr(true),
				},
				Data: &openrtb_ext.PriceFloorData{
					Currency: "USD",
					ModelGroups: []openrtb_ext.PriceFloorModelGroup{
						{
							ModelVersion: "3 dynamic model 1",
							Currency:     "USD",
							Values: map[string]float64{
								"banner|300x600|www.website5.com": 5,
								"*|*|*":                           7,
							},
							Schema: openrtb_ext.PriceFloorSchema{
								Fields:    []string{"mediaType", "size", "domain"},
								Delimiter: "|",
							},
						},
					},
				},
			},
		},
	}

	index := rand.Intn(maxFetchResults)
	return &sampleFloors[index]

	// Above code is added for testing purpose, shall be removed once sanity testing is done

	//	var fetchedResults fetchReult

	// Check for Rules in cache

	// fetch floors JSON
	//return fetchPriceFloorRules(account)
}

func fetchPriceFloorRules(account config.Account) *fetchReult {
	// If fetch is disabled
	fetchConfig := account.PriceFloors.Fetch
	if !fetchConfig.Enabled {
		return &fetchReult{
			fetchStatus: openrtb_ext.FetchNone,
		}
	}

	if !validator.IsURL(fetchConfig.URL) {
		return &fetchReult{
			fetchStatus: openrtb_ext.FetchError,
		}
	}

	_, fetchInprogress := fetchInProgress[fetchConfig.URL]
	if !fetchInprogress {
		fetchPriceFloorRulesAsynchronous(account)
	}

	// Rules not present in cache, fetch rules asynchronously
	return &fetchReult{
		fetchStatus: openrtb_ext.FetchInprogress,
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
