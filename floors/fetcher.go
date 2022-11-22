package floors

import (
	"container/heap"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/alitto/pond"
	validator "github.com/asaskevich/govalidator"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type FloorFetcher interface {
	Fetch(configs config.AccountPriceFloors) *openrtb_ext.PriceFloorRules
}

type PriceFloorFetcher struct {
	pool                *pond.WorkerPool
	fetchQueue          FetchQueue
	floorFetcherChannel chan FetchInfo
	done                chan struct{}
}

func NewPriceFloorFetcher(maxWorkers, maxCapacity int) *PriceFloorFetcher {

	floorFetcher := PriceFloorFetcher{
		pool:                pond.New(maxWorkers, maxCapacity),
		fetchQueue:          make(FetchQueue, 0, 100),
		floorFetcherChannel: make(chan FetchInfo, 1000),
		done:                make(chan struct{}),
	}

	go floorFetcher.priceFloorFetcher()

	return &floorFetcher
}

func (f *PriceFloorFetcher) Fetch(configs config.AccountPriceFloors) *openrtb_ext.PriceFloorRules {

	//check in cache: hit/miss
	//hit: directly return
	// pwd, _ := os.Getwd()
	// path := filepath.Join(pwd, "floor.json")
	// if _, err := os.Stat(path); err == nil {
	// 	content, err := ioutil.ReadFile("floor.json")
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	var data openrtb_ext.PriceFloorRules
	// 	err = json.Unmarshal(content, &data)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	return &data
	// }

	//miss: push to channel to fetch and return empty response
	if configs.Enabled && configs.Fetch.Enabled && len(configs.Fetch.URL) > 0 && validator.IsURL(configs.Fetch.URL) && configs.Fetch.Timeout > 0 {
		fetchInfo := FetchInfo{AccountFloorFetch: configs.Fetch, FetchPeriod: time.Now().Unix()}
		f.floorFetcherChannel <- fetchInfo
	}

	return nil
}

func (f *PriceFloorFetcher) floorfetcherWorker(configs config.AccountFloorFetch) {

	floorData := floorFetcherAndValidator(configs)
	if floorData != nil {
		// Update cache with new floor rules
		glog.Info("Updating Value in cache")
		// pwd, _ := os.Getwd()
		// content, err := json.Marshal(floorData)
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// path := filepath.Join(pwd, "floor.json")
		// err = ioutil.WriteFile(path, content, 0644)
		// if err != nil {
		// 	fmt.Println(err)
		// }
	}

	// Send to refetch channel
	f.floorFetcherChannel <- FetchInfo{AccountFloorFetch: configs, FetchPeriod: time.Now().Add(time.Duration(configs.Period) * time.Second).Unix()}

}

func (f *PriceFloorFetcher) Stop() {
	close(f.done)
}

func (f *PriceFloorFetcher) priceFloorFetcher() {

	//Create Ticker of 5 seconds
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case fetchInfo := <-f.floorFetcherChannel:
			// TODO :: Add check to get if URL is present in cache
			heap.Push(&f.fetchQueue, &fetchInfo)
		case <-ticker.C:
			currentTime := time.Now().Unix()
			for top := f.fetchQueue.Top(); top != nil && top.FetchPeriod < currentTime; top = f.fetchQueue.Top() {
				nextFetch := heap.Pop(&f.fetchQueue)
				status := f.pool.TrySubmit(func() {
					f.floorfetcherWorker(nextFetch.(*FetchInfo).AccountFloorFetch)
				})
				if !status {
					heap.Push(&f.fetchQueue, &nextFetch)
				}
			}
		case <-f.done:
			glog.Info("Price Floor fetcher terminated")
		}
	}
}

func floorFetcherAndValidator(configs config.AccountFloorFetch) *openrtb_ext.PriceFloorRules {

	floorResp, err := fetchPriceFloorRulesFromURL(configs.URL, configs.Timeout)
	if err != nil {
		glog.Errorf("Error while fetching floor data from URL: %s, reason : %s", configs.URL, err.Error())
		return nil
	}

	if len(floorResp) > configs.MaxFileSize {
		glog.Errorf("Recieved invalid floor data from URL: %s, reason : floor file size is greater than MaxFileSize", configs.URL)
		return nil
	}

	var priceFloors openrtb_ext.PriceFloorRules
	if err = json.Unmarshal(floorResp, &priceFloors); err != nil {
		glog.Errorf("Recieved invalid price floor json from URL: %s", configs.URL)
		return nil
	} else {
		err := validatePriceFloorRules(configs, &priceFloors)
		if err != nil {
			glog.Errorf("Validation failed for floor JSON from URL: %s, reason: %s", configs.URL, err.Error())
			return nil
		}
	}

	return &priceFloors
}

func fetchPriceFloorRulesFromURL(URL string, timeout int) ([]byte, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, nil)
	if err != nil {
		return nil, errors.New("error while forming http fetch request : " + err.Error())
	}

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, errors.New("error while getting response from url : " + err.Error())
	}

	if httpResp.StatusCode != 200 {
		return nil, errors.New("no response from server")
	}

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, errors.New("unable to read response")
	}
	defer httpResp.Body.Close()

	return respBody, nil
}

func validatePriceFloorRules(configs config.AccountFloorFetch, priceFloors *openrtb_ext.PriceFloorRules) error {

	if priceFloors.Data == nil {
		return errors.New("empty data in floor JSON")
	}

	if len(priceFloors.Data.ModelGroups) < 1 {
		return errors.New("no model groups found in price floor data")
	}

	for _, modelGroup := range priceFloors.Data.ModelGroups {
		if len(modelGroup.Values) < 1 || len(modelGroup.Values) > configs.MaxRules {
			return errors.New("invalid number of floor rules, floor rules should be greater than zero and less than MaxRules specified in account config")
		}

		if modelGroup.ModelWeight != nil {
			if *modelGroup.ModelWeight < 1 || *modelGroup.ModelWeight > 100 {
				return errors.New("modelGroup[].modelWeight should be greater than or equal to 1 and less than 100")
			}
		}

		if modelGroup.SkipRate < 0 || modelGroup.SkipRate > 100 {
			return errors.New("skip rate should be greater than or equal to 0 and less than 100")
		}

		if modelGroup.Default < 0 {
			return errors.New("modelGroup.Default should be greater than 0")
		}
	}

	return nil
}
