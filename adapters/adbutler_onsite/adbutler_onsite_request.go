package adbutler_onsite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
)

type AdButlerOnsiteRequest struct {
	ID          int                    `json:"ID,omitempty"`
	Size        string                 `json:"size,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Ads         string                 `json:"ads,omitempty"`
	KeyWords    []string               `json:"kw,omitempty"`
	ZoneIDs     []int                  `json:"zoneIDs,omitempty"`
	Limit       map[int]int            `json:"limit,omitempty"`
	Target      map[string]interface{} `json:"_abdk_json,omitempty"`
	UserID      string                 `json:"adb_uid,omitempty"`
	IP          string                 `json:"ip,omitempty"`
	UserAgent   string                 `json:"ua,omitempty"`
	Referrer    string                 `json:"referrer,omitempty"`
	PageID		int			           `json:"pid,omitempty"`
	Sequence    int			           `json:"place,omitempty"`
}

// getSimpleHash generates a simple hash for a given page name
func getSimpleHash(pageName string) int {
	const primeBase = 31
	const mod = 1e9 + 9 // A large prime number for modulus

	hashValue := 0
	for _, char := range pageName {
		hashValue = (hashValue*primeBase + int(char)) % mod
	}

	return hashValue
}

/*// Function to map ad type integers to their string representations
func adTypeToString(adType int) string {
	switch adType {
	case RequestAdtype_Banner:
		return DBAdtype_Banner
	case RequestAdtype_Custom_Html:
		return DBAdtype_Custom_Html
	default:
		return ""
	}
}

// Function to find matching ad types and return specific strings based on the type
func findMatchingAdTypes(requestedAdTypes []int, actualAdTypes string) []string {
	// Split the actual ad types string into a slice of strings
	actualAdTypesSlice := strings.Split(actualAdTypes, ",")

	// Create a map for quick lookup of actual ad types
	actualAdTypesMap := make(map[string]bool)
	for _, adType := range actualAdTypesSlice {
		actualAdTypesMap[strings.TrimSpace(adType)] = true
	}

	// Find matching ad types and append specific strings
	var matchingAdTypes []string
	for _, requestedAdType := range requestedAdTypes {
		adTypeStr := adTypeToString(requestedAdType)
		if actualAdTypesMap[adTypeStr] {
			if adTypeStr == DBAdtype_Banner {
				matchingAdTypes = append(matchingAdTypes, AdButlerAdtype_Banner)
			} else if adTypeStr == DBAdtype_Custom_Html {
				matchingAdTypes = append(matchingAdTypes, AdButlerAdtype_Custom_Html)
			}
		}
	}

	return matchingAdTypes
}*/

func isInventorySizeMatch(inventory openrtb_ext.CMOnsiteInventoryDetails, banner *openrtb2.Banner) bool {
	if banner == nil {
		return false
	}

	if(inventory.Width == 0 || inventory.Height == 0) {
		return true
	}

	for _, format := range banner.Format {
		if format.W == inventory.Width && format.H == inventory.Height {
			return true
		}
	}
	return false
}


func getInventoryTypes( actualAdTypes string) []string {
	// Split the actual ad types string into a slice of strings
	actualAdTypesSlice := strings.Split(actualAdTypes, ",")

	var matchingAdTypes []string
	for _, adType := range actualAdTypesSlice {
		if strings.TrimSpace(adType) == DBAdtype_Banner {
			matchingAdTypes = append(matchingAdTypes, AdButlerAdtype_Banner)
		} else {
			matchingAdTypes = append(matchingAdTypes, AdButlerAdtype_Custom_Html)
		}
	}
	return matchingAdTypes
}

func (a *AdButlerOnsiteAdapter) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	siteExt, requestExt, errors := adapters.ValidateCMOnsiteRequest(request)
	if len(errors) > 0 {
		return nil, errors
	}

	if siteExt == nil || requestExt == nil {
		return nil, []error{&errortypes.BadInput{
			Message: "Missing required ext fields which contains inventory details",
		}}
	}

	inventoryDetails, accountID, _:= adapters.GetInventoryAndAccountDetailsCMOnsite(requestExt)
	
	if inventoryDetails == nil || accountID == "" {
		return nil, []error{&errortypes.BadInput{
			Message: "Missing inventory details or accountID details",
		}}
	}
	
	var adButlerReq AdButlerOnsiteRequest

	// Convert accountID to an integer
	id, err := strconv.Atoi(accountID)
	if err != nil {
		return nil, []error{&errortypes.BadInput{
			Message: "accountID details is not Valid",
		}}
	}

	// Assign the converted integer to adButlerReq.ID
	adButlerReq.ID = id
	adButlerReq.Type = AdButler_Req_Type
	adButlerReq.Ads = AdButler_Req_Ads

    adButlerReq.Target = make(map[string]interface{})

	//Add Geo Targeting
	if request.Device != nil && request.Device.Geo != nil {
		if request.Device.Geo.Country != "" {
			adButlerReq.Target[adapters.COUNTRY] = request.Device.Geo.Country
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

	//Add Dynamic Targeting from AdRequest
	for _, targetObj := range requestExt.Targeting {
		key := targetObj.Name
		adButlerReq.Target[key] = targetObj.Value
	}

	adButlerReq.Sequence = requestExt.Sequence
	adButlerReq.PageID = getSimpleHash(siteExt.Page)

	limitMap := make(map[int]int)
	appendedZoneIDs := make(map[int]bool)
	for _, imp := range request.Imp {
		// Parse each imp element here
		inventory ,ok := inventoryDetails[InventoryIDOnsite_Prefix + imp.TagID]
		if !ok {
			continue
		}
		adButlerAdtypes := getInventoryTypes(inventory.Adtype)
		isSizeMatch := isInventorySizeMatch(inventory, imp.Banner)
		if(len(adButlerAdtypes) > 0 && isSizeMatch) {
			// Check if the ZoneID has already been appended
			if !appendedZoneIDs[inventory.AdbulterZoneID] {
				adButlerReq.ZoneIDs = append(adButlerReq.ZoneIDs, inventory.AdbulterZoneID)
				appendedZoneIDs[inventory.AdbulterZoneID] = true
				limitMap[inventory.AdbulterZoneID] += 1
			} else {
				limitMap[inventory.AdbulterZoneID] += 1
			}
		}
	}
	
	adButlerReq.Limit = limitMap
	
	reqJSON, err := json.Marshal(adButlerReq)
	if err != nil {
		return nil, []error{err}
	}

	// Pretty-print the JSON request body for debugging
	var prettyReqJSON bytes.Buffer
	err = json.Indent(&prettyReqJSON, reqJSON, "", "  ")
	if err != nil {
		fmt.Println("Failed to parse JSON:", err)
		return nil, []error{err}
	}
	fmt.Println("Request Body:")
	fmt.Println(prettyReqJSON.String())

	headers := http.Header{}
	headers.Add("Content-Type", "application/json")

	return []*adapters.RequestData{{
		Method:  "POST",
		Uri:     a.endpoint,
		Body:    reqJSON,
		Headers: headers,
	}}, nil

}







