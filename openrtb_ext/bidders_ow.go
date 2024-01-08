package openrtb_ext

// import (
// 	"encoding/json"
// 	"fmt"

// 	"github.com/prebid/prebid-server/adapters/rtbbidder"
// )

// // func SyncRTBBidders() error {
// // 	// list of rtb bidders from wrapper_partner
// // 	var rtbBidders []BidderName
// // 	rtbBidders = append(rtbBidders, BidderName("myrrtbbidder"))
// // 	for _, rtbBidder := range rtbBidders {
// // 		SetAliasBidderName(string(rtbBidder), rtbBidder)
// // 	}
// // 	fmt.Println(CoreBidderNames())
// // 	return nil
// // }

// func NewOpenWrapBidderParamsValidator(schemaDirectory string) (BidderParamValidator, error) {
// 	bidderParamsValidator, err := NewBidderParamsValidator(schemaDirectory)
// 	rtbBidder := rtbbidder.GetInstance()
// 	owBidderParamsValidator := &OpenWrapbidderParamValidator{
// 		prebidValidator: bidderParamsValidator,
// 		schemaDirectory: schemaDirectory + "/rtb",
// 	}
// 	rtbBidder.Syncher.SyncBiddersParameters(owBidderParamsValidator)
// 	return owBidderParamsValidator, err
// }

// type OpenWrapbidderParamValidator struct {
// 	prebidValidator BidderParamValidator
// 	rtbValidator    BidderParamValidator
// 	schemaDirectory string
// }

// // LoadSchema reloads all RTB bidder's bidder-params json
// func (owv *OpenWrapbidderParamValidator) LoadSchema() {
// 	validator, err := NewBidderParamsValidator(owv.schemaDirectory)
// 	if err == nil {
// 		owv.rtbValidator = validator
// 	}
// }

// // Schema implements BidderParamValidator.
// func (owv *OpenWrapbidderParamValidator) Schema(name BidderName) string {
// 	schema := owv.prebidValidator.Schema(name)
// 	if schema == "" {
// 		owv.rtbValidator.Schema(name)
// 	}
// 	return schema
// }

// // Validate implements BidderParamValidator.
// func (owValidator *OpenWrapbidderParamValidator) Validate(name BidderName, ext json.RawMessage) error {
// 	if name == BidderRTBBidder {
// 		// syncer := rtbbidder.GetInstance().Syncher
// 		fmt.Printf("Validator bidder params for %s", name)
// 		return nil
// 	}
// 	return owValidator.prebidValidator.Validate(name, ext)
// }
