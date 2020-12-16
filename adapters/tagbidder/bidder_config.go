package tagbidder

import (
	"encoding/json"
	"io/ioutil"

	"github.com/golang/glog"
)

//Flags of each tag bidder
type Flags struct {
	RemoveEmptyParam bool `json:"remove_empty,omitempty"`
}

//Keys each macro mapping key definition
type Keys struct {
	Cached    *bool     `json:"cached,omitempty"`
	Value     string    `json:"value,omitempty"`
	ValueType ValueType `json:"type,omitempty"`
}

//BidderConfig mapper json
type BidderConfig struct {
	URL          string              `json:"url,omitempty"`
	ResponseType ResponseHandlerType `json:"response,omitempty"`
	Flags        Flags               `json:"flags,omitempty"`
	Keys         map[string]Keys     `json:"keys,omitempty"`
}

var bidderConfig map[string]*BidderConfig

//RegisterBidderConfig will be used by each bidder to set its respective macro Mapper
func RegisterBidderConfig(bidder string, config *BidderConfig) {
	bidderConfig[bidder] = config
}

//GetBidderConfig will return Mapper of specific bidder
func GetBidderConfig(bidder string) *BidderConfig {
	return bidderConfig[bidder]
}

//FetchBidderConfig returns new Mapper from JSON details
func FetchBidderConfig(confDir string, bidders []string) {
	for _, bidderName := range bidders {
		bidderString := string(bidderName)
		fileData, err := ioutil.ReadFile(confDir + "/" + bidderString + ".json")
		if err != nil {
			glog.Fatalf("error reading from file %s: %v", confDir+"/"+bidderString+".json", err)
		}

		var bidderConfig BidderConfig
		if err := json.Unmarshal([]byte(fileData), &bidderConfig); nil != err {
			glog.Fatalf("error parsing json in file %s: %v", confDir+"/"+bidderString+".json", err)
		}
		RegisterBidderConfig(bidderString, &bidderConfig)

		mapper := NewMapperFromConfig(&bidderConfig)
		if nil == mapper {
			glog.Fatalf("no query parameters mapper for bidder " + bidderString)
		}
		RegisterBidderMapper(bidderString, mapper)
	}
}
