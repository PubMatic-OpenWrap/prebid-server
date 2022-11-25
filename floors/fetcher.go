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
	"github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

const (
	CACHE_EXPIRY_ROUTINE_RUN_INTERVAL = 60 * time.Minute
	CACHE_DEFAULT_EXPIRY_INTERVAL     = 360 * time.Minute
)

type FloorFetcher interface {
	Fetch(configs config.AccountPriceFloors) (*openrtb_ext.PriceFloorRules, string)
}

type PriceFloorFetcher struct {
	pool            *pond.WorkerPool // Goroutines worker pool
	fetchQueue      FetchQueue       // Priority Queue to fetch floor data
	fetchInprogress map[string]bool  // Map of URL with fetch status
	configReceiver  chan FetchInfo   // Channel which recieves URLs to be fetched
	done            chan struct{}    // Channel to close fetcher
	cache           *cache.Cache     // cache
}

type FetchInfo struct {
	config.AccountFloorFetch
	FetchTime      int64
	RefetchRequest bool
}

type FetchQueue []*FetchInfo

func (fq FetchQueue) Len() int {
	return len(fq)
}

func (fq FetchQueue) Less(i, j int) bool {
	return fq[i].FetchTime < fq[j].FetchTime
}

func (fq FetchQueue) Swap(i, j int) {
	fq[i], fq[j] = fq[j], fq[i]
}

func (fq *FetchQueue) Push(element interface{}) {
	fetchInfo := element.(*FetchInfo)
	*fq = append(*fq, fetchInfo)
}

func (fq *FetchQueue) Pop() interface{} {
	old := *fq
	n := len(old)
	fetchInfo := old[n-1]
	old[n-1] = nil // avoid memory leak
	*fq = old[0 : n-1]
	return fetchInfo
}

func (fq *FetchQueue) Top() *FetchInfo {
	old := *fq
	if len(old) == 0 {
		return nil
	}
	return old[0]
}

func NewPriceFloorFetcher(maxWorkers, maxCapacity int) *PriceFloorFetcher {

	floorFetcher := PriceFloorFetcher{
		pool:            pond.New(maxWorkers, maxCapacity),
		fetchQueue:      make(FetchQueue, 0, 100),
		fetchInprogress: make(map[string]bool),
		configReceiver:  make(chan FetchInfo, maxCapacity),
		done:            make(chan struct{}),
		cache:           cache.New(CACHE_DEFAULT_EXPIRY_INTERVAL, CACHE_EXPIRY_ROUTINE_RUN_INTERVAL),
	}

	go floorFetcher.Fetcher()

	return &floorFetcher
}

func (f *PriceFloorFetcher) SetWithExpiry(key string, value interface{}, expiry time.Duration) {
	f.cache.Set(key, value, expiry)
}

func (f *PriceFloorFetcher) Set(key string, value interface{}) {
	f.cache.Set(key, value, CACHE_DEFAULT_EXPIRY_INTERVAL)
}

func (f *PriceFloorFetcher) Get(key string) (interface{}, bool) {
	return f.cache.Get(key)
}

func (f *PriceFloorFetcher) Fetch(configs config.AccountPriceFloors) (*openrtb_ext.PriceFloorRules, string) {

	// Check for floors JSON in cache
	var fetcheRes openrtb_ext.PriceFloorRules
	result, ret := f.Get(configs.Fetch.URL)
	if ret {
		fetcheRes = result.(openrtb_ext.PriceFloorRules)
		if fetcheRes.Data != nil {
			return &fetcheRes, openrtb_ext.FetchSuccess
		} else {
			return nil, openrtb_ext.FetchError
		}
	}

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
	// 	return &data, openrtb_ext.FetchSuccess
	// }

	//miss: push to channel to fetch and return empty response
	if configs.Enabled && configs.Fetch.Enabled && len(configs.Fetch.URL) > 0 && validator.IsURL(configs.Fetch.URL) && configs.Fetch.Timeout > 0 {
		fetchInfo := FetchInfo{AccountFloorFetch: configs.Fetch, FetchTime: time.Now().Unix(), RefetchRequest: false}
		f.configReceiver <- fetchInfo
	}

	return nil, openrtb_ext.FetchInprogress
}

func (f *PriceFloorFetcher) worker(configs config.AccountFloorFetch) {

	floorData := fetchAndValidate(configs)
	if floorData != nil {
		// Update cache with new floor rules
		glog.Info("Updating Value in cache")
		f.Set(configs.URL, floorData)
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
	f.configReceiver <- FetchInfo{AccountFloorFetch: configs, FetchTime: time.Now().Add(time.Duration(configs.Period) * time.Second).Unix(), RefetchRequest: true}

}

func (f *PriceFloorFetcher) Stop() {
	close(f.done)
}

func (f *PriceFloorFetcher) Fetcher() {

	//Create Ticker of 5 seconds
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case fetchInfo := <-f.configReceiver:
			if fetchInfo.RefetchRequest {
				heap.Push(&f.fetchQueue, &fetchInfo)
			} else {
				if _, ok := f.fetchInprogress[fetchInfo.URL]; !ok {
					f.fetchInprogress[fetchInfo.URL] = true
					heap.Push(&f.fetchQueue, &fetchInfo)
				}
			}
		case <-ticker.C:
			currentTime := time.Now().Unix()
			for top := f.fetchQueue.Top(); top != nil && top.FetchTime < currentTime; top = f.fetchQueue.Top() {
				nextFetch := heap.Pop(&f.fetchQueue)
				status := f.pool.TrySubmit(func() {
					f.worker(nextFetch.(*FetchInfo).AccountFloorFetch)
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

func fetchAndValidate(configs config.AccountFloorFetch) *openrtb_ext.PriceFloorRules {

	floorResp, err := fetchFloorRulesFromURL(configs.URL, configs.Timeout)
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
		err := validateRules(configs, &priceFloors)
		if err != nil {
			glog.Errorf("Validation failed for floor JSON from URL: %s, reason: %s", configs.URL, err.Error())
			return nil
		}
	}

	return &priceFloors
}

func fetchFloorRulesFromURL(URL string, timeout int) ([]byte, error) {

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

func validateRules(configs config.AccountFloorFetch, priceFloors *openrtb_ext.PriceFloorRules) error {

	if priceFloors.Data == nil {
		return errors.New("empty data in floor JSON")
	}

	if len(priceFloors.Data.ModelGroups) == 0 {
		return errors.New("no model groups found in price floor data")
	}

	for _, modelGroup := range priceFloors.Data.ModelGroups {
		if len(modelGroup.Values) == 0 || len(modelGroup.Values) > configs.MaxRules {
			return errors.New("invalid number of floor rules, floor rules should be greater than zero and less than MaxRules specified in account config")
		}

		if modelGroup.ModelWeight != nil && (*modelGroup.ModelWeight < 1 || *modelGroup.ModelWeight > 100) {
			return errors.New("modelGroup[].modelWeight should be greater than or equal to 1 and less than 100")
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
