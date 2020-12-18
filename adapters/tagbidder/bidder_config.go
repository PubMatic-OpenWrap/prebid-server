package tagbidder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
	"github.com/golang/glog"
)

//Flags of each tag bidder
type Flags struct {
	RemoveEmptyParam bool `json:"remove_empty,omitempty"`
}

//Keys each macro mapping key definition
type Keys struct {
	Cached    *bool          `json:"cached,omitempty"`
	Value     string         `json:"value,omitempty"`
	ValueType MacroValueType `json:"type,omitempty"`
}

//BidderConfig mapper json
type BidderConfig struct {
	URL          string              `json:"url,omitempty"`
	ResponseType ResponseHandlerType `json:"response,omitempty"`
	Flags        Flags               `json:"flags,omitempty"`
	Keys         map[string]Keys     `json:"keys,omitempty"`
}

var bidderConfig = map[string]*BidderConfig{}

//RegisterBidderConfig will be used by each bidder to set its respective macro Mapper
func RegisterBidderConfig(bidder string, config *BidderConfig) {
	bidderConfig[bidder] = config
}

//GetBidderConfig will return Mapper of specific bidder
func GetBidderConfig(bidder string) *BidderConfig {
	return bidderConfig[bidder]
}

//InitTagBidderConfig returns new Mapper from JSON details
func InitTagBidderConfig(schemaDirectory string, tagBidderMap map[string]openrtb_ext.BidderName) error {
	fileInfos, err := ioutil.ReadDir(schemaDirectory)
	if err != nil {
		return fmt.Errorf("Failed to read JSON schemas from directory %s. %v", schemaDirectory, err)
	}

	for _, fileInfo := range fileInfos {

		//checking for invalid tag bidder names
		bidderName := strings.TrimSuffix(fileInfo.Name(), ".json")
		if _, isValid := tagBidderMap[bidderName]; !isValid {
			return fmt.Errorf("File %s/%s does not match a valid BidderName", schemaDirectory, fileInfo.Name())
		}

		//bidder config file absolute path
		toOpen, err := filepath.Abs(filepath.Join(schemaDirectory, fileInfo.Name()))
		if err != nil {
			return fmt.Errorf("Failed to get an absolute representation of the path: %s, %v", toOpen, err)
		}

		//reading config file contents
		fileBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", schemaDirectory, fileInfo.Name()))
		if err != nil {
			return fmt.Errorf("Failed to read file %s/%s: %v", schemaDirectory, fileInfo.Name(), err)
		}

		//reading bidder config values
		var bidderConfig BidderConfig
		if err := json.Unmarshal(fileBytes, &bidderConfig); nil != err {
			glog.Fatalf("error parsing json in file %s: %v", schemaDirectory+"/"+bidderName+".json", err)
		}

		//reading its tag parameter mapper from config
		mapper := NewMapperFromConfig(&bidderConfig)
		if nil == mapper {
			glog.Fatalf("no query parameters mapper for bidder " + bidderName)
		}

		//register tag bidder configurations
		RegisterBidderConfig(bidderName, &bidderConfig)
		RegisterBidderMapper(bidderName, mapper)
	}
	return nil
}
