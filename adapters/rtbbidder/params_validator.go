package rtbbidder

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/prebid/prebid-server/openrtb_ext"
)

// func SyncRTBBidders() error {
// 	// list of rtb bidders from wrapper_partner
// 	var rtbBidders []BidderName
// 	rtbBidders = append(rtbBidders, BidderName("myrrtbbidder"))
// 	for _, rtbBidder := range rtbBidders {
// 		SetAliasBidderName(string(rtbBidder), rtbBidder)
// 	}
// 	fmt.Println(CoreBidderNames())
// 	return nil
// }

func NewOpenWrapBidderParamsValidator(schemaDirectory string, bidderParamsValidator openrtb_ext.BidderParamValidator) (openrtb_ext.BidderParamValidator, error) {
	// bidderParamsValidator, err := openrtb_ext.NewBidderParamsValidator(schemaDirectory)
	rtbBidder := getInstance()
	owBidderParamsValidator := &OpenWrapbidderParamValidator{
		prebidValidator: bidderParamsValidator,
		schemaDirectory: schemaDirectory + rtbBidder.syncher.syncPath,
	}
	owBidderParamsValidator.rtbValidator, _ = openrtb_ext.NewBidderParamsValidator(owBidderParamsValidator.schemaDirectory)
	if rtbBidder == nil {
		return nil, errors.New("rtbbidder instance is not initialized")
	}
	rtbBidder.syncher.syncBiddersParameters(owBidderParamsValidator)
	return owBidderParamsValidator, nil
}

type OpenWrapbidderParamValidator struct {
	prebidValidator openrtb_ext.BidderParamValidator
	rtbValidator    openrtb_ext.BidderParamValidator
	schemaDirectory string
}

// LoadSchema reloads all RTB bidder's bidder-params json
func (owv *OpenWrapbidderParamValidator) LoadSchema() []string {
	validator, err := openrtb_ext.NewBidderParamsValidator(owv.schemaDirectory)
	if err == nil {
		owv.rtbValidator = validator
		fileInfos, err := os.ReadDir(owv.schemaDirectory)
		if err == nil {
			bidders := make([]string, 0)
			for _, fileInfo := range fileInfos {
				bidders = append(bidders, fileInfo.Name())
			}
			return bidders
		}
	}
	return []string{}
}

// Schema implements BidderParamValidator.
func (owv *OpenWrapbidderParamValidator) Schema(name openrtb_ext.BidderName) string {
	schema := owv.prebidValidator.Schema(name)
	if schema == "" {
		schema = owv.rtbValidator.Schema(name)
	}
	return schema
}

// Validate implements BidderParamValidator.
func (owValidator *OpenWrapbidderParamValidator) Validate(name openrtb_ext.BidderName, ext json.RawMessage) error {
	rtbBidder := getInstance()
	if _, ok := rtbBidder.syncher.syncedBiddersMap[string(name)]; ok || name == openrtb_ext.BidderRTBBidder {
		fmt.Printf("Validator bidder params for %s", name)
		return owValidator.rtbValidator.Validate(name, ext)
	}

	return owValidator.prebidValidator.Validate(name, ext)
}
