package adbutler_sponsored

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
)

type AdButlerSponsoredRequest struct {
	SearchString  string                 `json:"search,omitempty"`
	SearchType    string                 `json:"search_type,omitempty"`
	Params        map[string][]string    `json:"params,omitempty"`
	Identifiers   []string               `json:"identifiers,omitempty"`
	Target        map[string]interface{} `json:"_abdk_json,omitempty"`
	Limit         int                    `json:"limit,omitempty"`
	Source        string                 `json:"source,omitempty"`
	UserID        string                 `json:"adb_uid,omitempty"`
	IP            string                 `json:"ip,omitempty"`
	UserAgent     string                 `json:"ua,omitempty"`
	Referrer      string                 `json:"referrer,omitempty"`
	FloorCPC      float64                `json:"bid_floor_cpc,omitempty"`
	IsTestRequest bool                   `json:"test_request,omitempty"`
}

func isLowercaseNumbersDashes(s string) bool {
	// Define a regular expression pattern to match lowercase letters, numbers, and dashes
	pattern := "^[a-z0-9-]+$"
	re := regexp.MustCompile(pattern)

	// Use the MatchString function to check if the string matches the pattern
	return re.MatchString(s)
}

func (a *AdButlerSponsoredAdapter) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {

	commerceExt, siteExt, _, errors := adapters.ValidateCommRequest(request)
	if len(errors) > 0 {
		return nil, errors
	}

	var configValueMap = make(map[string]string)
	var configTypeMap = make(map[string]int)
	for _, obj := range commerceExt.Bidder.CustomConfig {
		configValueMap[obj.Key] = obj.Value
		configTypeMap[obj.Key] = obj.Type
	}

	var adButlerReq AdButlerSponsoredRequest
	//Assign Page Source if Present
	if siteExt != nil {
		if isLowercaseNumbersDashes(siteExt.Page) {
			adButlerReq.Source = siteExt.Page
		}
	}

	//Retrieve AccountID and ZoneID from Request and Build endpoint Url
	var accountID, zoneID string
	val, ok := configValueMap[adapters.BIDDERDETAILS_PREFIX+BD_ACCOUNT_ID]
	if ok {
		accountID = val
	}

	val, ok = configValueMap[adapters.BIDDERDETAILS_PREFIX+BD_ZONE_ID]
	if ok {
		zoneID = val
	}

	endPoint, err := a.buildEndpointURL(accountID, zoneID)
	if err != nil {
		return nil, []error{err}
	}

	adButlerReq.Target = make(map[string]interface{})
	//Add User Targeting
	if request.User != nil {
		if request.User.Yob > 0 {
			now := time.Now()
			age := int64(now.Year()) - request.User.Yob
			adButlerReq.Target[adapters.USER_AGE] = age
		}

		if request.User.Gender != "" {
			if strings.EqualFold(request.User.Gender, "M") {
				adButlerReq.Target[adapters.USER_GENDER] = adapters.GENDER_MALE
			} else if strings.EqualFold(request.User.Gender, "F") {
				adButlerReq.Target[adapters.USER_GENDER] = adapters.GENDER_FEMALE
			} else if strings.EqualFold(request.User.Gender, "O") {
				adButlerReq.Target[adapters.USER_GENDER] = adapters.GENDER_OTHER
			}
		}
	}

	//Add Geo Targeting
	if request.Device != nil && request.Device.Geo != nil {
		if request.Device.Geo.Country != "" {
			adButlerReq.Target[adapters.COUNTRY] = request.Device.Geo.Country
		}
		if request.Device.Geo.Region != "" {
			adButlerReq.Target[adapters.REGION] = request.Device.Geo.Region
		}
		if request.Device.Geo.City != "" {
			adButlerReq.Target[adapters.CITY] = request.Device.Geo.City
		}
	}
	//Add Geo Targeting
	if request.Device != nil {
		switch request.Device.DeviceType {
		case 1:
			adButlerReq.Target[adapters.DEVICE] = adapters.DEVICE_COMPUTER
		case 2:
			adButlerReq.Target[adapters.DEVICE] = adapters.DEVICE_PHONE
		case 3:
			adButlerReq.Target[adapters.DEVICE] = adapters.DEVICE_TABLET
		case 4:
			adButlerReq.Target[adapters.DEVICE] = adapters.DEVICE_CONNECTEDDEVICE
		}
	}

	//Add Page Source Targeting
	if adButlerReq.Source != "" {
		adButlerReq.Target[PAGE_SOURCE] = adButlerReq.Source
	}

	//Add Dynamic Targeting from AdRequest
	for _, targetObj := range commerceExt.ComParams.Targeting {
		key := targetObj.Name
		adButlerReq.Target[key] = targetObj.Value
	}
	//Add Identifiers from AdRequest
	for _, prefObj := range commerceExt.ComParams.Preferred {
		adButlerReq.Identifiers = append(adButlerReq.Identifiers, prefObj.ProductID)
	}

	if commerceExt.ComParams.Filtering != nil {
		subcategoryMap := make(map[string]bool)
		if value, ok := configValueMap[adapters.PRODUCTTEMPLATE_PREFIX+PD_TEMPLATE_SUBCATEGORY]; ok {
			subcategories := strings.Split(value, ProductTemplate_Separator)
			for _, subcategory := range subcategories {
				subcategoryMap[subcategory] = true
			}
			for _, category := range commerceExt.ComParams.Filtering {
				key := category.Name
				if _, ok := subcategoryMap[key]; !ok {
					errors = append(errors, &errortypes.InvalidProductFiltering{
						Message: "Invalid Subcategory : " + key,
					})
					return nil, errors
				}
			}
		}
	}

	//Add Category Params from AdRequest
	if len(adButlerReq.Identifiers) <= 0 && commerceExt.ComParams.Filtering != nil && len(commerceExt.ComParams.Filtering) > 0 {

		adButlerReq.Params = make(map[string][]string)
		for _, category := range commerceExt.ComParams.Filtering {
			key := category.Name
			value := category.Value
			adButlerReq.Params[key] = value
		}
	}

	//Assign Search Term if present along with searchType
	if len(adButlerReq.Identifiers) <= 0 && commerceExt.ComParams.Filtering == nil && commerceExt.ComParams.SearchTerm != "" {
		adButlerReq.SearchString = commerceExt.ComParams.SearchTerm
		if commerceExt.ComParams.SearchType == SEARCHTYPE_EXACT ||
			commerceExt.ComParams.SearchType == SEARCHTYPE_BROAD {
			adButlerReq.SearchType = commerceExt.ComParams.SearchType
		} else {
			val, ok := configValueMap[SEARCHTYPE]
			if ok {
				adButlerReq.SearchType = val
			} else {
				adButlerReq.SearchType = SEARCHTYPE_DEFAULT
			}
		}
	}

	adButlerReq.IP = request.Device.IP
	// Domain Name from Site Object if Prsent or App Obj
	if request.Site != nil {
		adButlerReq.Referrer = request.Site.Domain
	} else {
		adButlerReq.Referrer = request.App.Domain
	}

	// Take BidFloor from BidRequest - High Priority, Otherwise from Auction Config
	if request.Imp[0].BidFloor > 0 {
		adButlerReq.FloorCPC = request.Imp[0].BidFloor
	} else {
		val, ok := configValueMap[adapters.AUCTIONDETAILS_PREFIX+adapters.AD_FLOOR_PRICE]
		if ok {
			if floorPrice, err := strconv.ParseFloat(val, 64); err == nil {
				adButlerReq.FloorCPC = floorPrice
			}
		}
	}

	//Test Request
	if commerceExt.ComParams.TestRequest {
		adButlerReq.IsTestRequest = true
	}
	adButlerReq.UserID = request.User.ID
	adButlerReq.UserAgent = request.Device.UA
	adButlerReq.Limit = commerceExt.ComParams.SlotsRequested

	//Temporarily for Debugging
	//u, _ := json.Marshal(adButlerReq)
	//fmt.Println(string(u))

	reqJSON, err := json.Marshal(adButlerReq)
	if err != nil {
		return nil, []error{err}
	}

	headers := http.Header{}
	headers.Add("Content-Type", "application/json")

	return []*adapters.RequestData{{
		Method:  "POST",
		Uri:     endPoint,
		Body:    reqJSON,
		Headers: headers,
	}}, nil

}

