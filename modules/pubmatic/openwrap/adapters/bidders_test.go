package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func init() {
	InitBidders("../../../../static/bidder-params")
}

func AssertJSON(t *testing.T, expectedJSON, actualJSON json.RawMessage, msgAndArgs ...interface{}) bool {
	msgAndArgs = append(msgAndArgs, fmt.Sprintf("expectedJSON: %s\nactualJSON: %s", string(expectedJSON), string(actualJSON)))
	//assert.JSONq returns '('') invalid json' for empty string, to avoid these cases if actual and expected have equal zero lengths, we are returing as valid assertion
	if len(expectedJSON) == 0 && len(actualJSON) == 0 {
		return true
	}
	return assert.JSONEq(t, string(expectedJSON), string(actualJSON), msgAndArgs...)
}

func AssertJSONObject(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	expectedJSON, _ := json.Marshal(expected)
	actualJSON, _ := json.Marshal(actual)
	return AssertJSON(t, expectedJSON, actualJSON, msgAndArgs...)
}

func GetJSON(obj interface{}) string {
	data, _ := json.Marshal(obj)
	return string(data[:])
}

func readTestCasesFromFile(t *testing.T, filePath string, tests interface{}) {
	//reading testcases from file
	testData, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	err = json.Unmarshal(testData, tests)
	assert.NoError(t, err)
}

// getPrebidBidderParamsValidator getting prebid bidder params validator
func getPrebidBidderParamsValidator(t *testing.T, schemaDirectory string) openrtb_ext.BidderParamValidator {
	validator, err := openrtb_ext.NewBidderParamsValidator(schemaDirectory)
	if err != nil {
		t.Logf("failed in getPrebidBidderParamsValidator, schemaDirectory:%s, error:%s", schemaDirectory, err.Error())
		t.FailNow()
	}
	return validator
}

func formBidderKeywordsMap() map[string]*models.BidderExtension {
	kvp1 := formKeyVal("key1", []string{"val1", "val2"})
	kvp2 := formKeyVal("key2", []string{"val3", "val4"})
	kvArr := []models.KeyVal{kvp1, kvp2}
	bmap := map[string]*models.BidderExtension{
		"appnexus": {KeyWords: kvArr},
		"pubmatic": {KeyWords: kvArr},
	}
	return bmap
}

func formKeyVal(key string, values []string) models.KeyVal {
	kv := models.KeyVal{
		Key:    key,
		Values: values,
	}
	return kv
}

func formBidderKeywordsMapForAppnexusAlias() map[string]*models.BidderExtension {
	kvp1 := formKeyVal("key1", []string{"val1", "val2"})
	kvp2 := formKeyVal("key2", []string{"val3", "val4"})
	kvArr := []models.KeyVal{kvp1, kvp2}
	bmap := map[string]*models.BidderExtension{
		"appnexus-alias": {KeyWords: kvArr},
		"pubmatic":       {KeyWords: kvArr},
	}
	return bmap
}

func TestPrepareBidParamJSONForPartner33across(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name: "All params present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"siteId":    "testSite",
					"productId": "testProduct",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.Bidder33Across),
				bidderCode:  string(openrtb_ext.Bidder33Across),
			},
			want:    json.RawMessage(`{"productId":"testProduct","siteId":"testSite"}`),
			wantErr: false,
		},
		{
			name: "siteId is missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"productId": "testProduct",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.Bidder33Across),
				bidderCode:  string(openrtb_ext.Bidder33Across),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "productId is missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"siteId": "testSite",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.Bidder33Across),
				bidderCode:  string(openrtb_ext.Bidder33Across),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "required params are missing",
			args: args{
				reqID:       "",
				width:       nil,
				height:      nil,
				fieldMap:    map[string]interface{}{},
				slotKey:     "",
				adapterName: string(openrtb_ext.Bidder33Across),
				bidderCode:  string(openrtb_ext.Bidder33Across),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			assert.Equal(t, tt.wantErr, err != nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerAdf(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All params present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"mid": "1234",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderAdf),
			},
			want: json.RawMessage(`{"mid":1234}`),
		},
		{
			name: "required param mid is missing",
			args: args{
				reqID:       "",
				width:       nil,
				height:      nil,
				fieldMap:    map[string]interface{}{},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderAdf),
			},
			want: nil,
		},
		{
			name: "required param mid is not an integer",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"mid": "abc",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderAdf),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, string(openrtb_ext.BidderAdf), nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerCriteo(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "Only zoneId present",
			args: args{
				fieldMap: map[string]interface{}{
					"zoneId": "1",
				},
				adapterName: string(openrtb_ext.BidderCriteo),
			},
			want: json.RawMessage(`{"zoneId":1}`),
		},
		{
			name: "Only networkId present",
			args: args{
				fieldMap: map[string]interface{}{
					"networkId": "4",
				},
				adapterName: string(openrtb_ext.BidderCriteo),
			},
			want: json.RawMessage(`{"networkId":4}`),
		},
		{
			name: "All params present",
			args: args{
				fieldMap: map[string]interface{}{
					"zoneId":    "1",
					"zoneid":    "2",
					"networkid": "3",
					"networkId": "4",
				},
				adapterName: string(openrtb_ext.BidderCriteo),
			},
			want: json.RawMessage(`{"zoneId":1}`),
		},
		{
			name: "No required params present",
			args: args{
				fieldMap: map[string]interface{}{
					"ZoneId": "5",
				},
				adapterName: string(openrtb_ext.BidderCriteo),
			},
			want: nil,
		},
		{
			name: "Invalid params value",
			args: args{
				fieldMap: map[string]interface{}{
					"zoneId": "a",
				},
				adapterName: string(openrtb_ext.BidderCriteo),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerAdform(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All params present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"mid": "1234",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderAdf),
			},
			want: json.RawMessage(`{"mid":1234}`),
		},
		{
			name: "required param mid is missing",
			args: args{
				reqID:       "",
				width:       nil,
				height:      nil,
				fieldMap:    map[string]interface{}{},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderAdf),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, string(openrtb_ext.BidderAdf), nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerSovrn(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All params present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"tagid":    "315045",
					"bidfloor": "0.04",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSovrn),
			},
			want: json.RawMessage(`{"tagid":"315045","bidfloor":0.04}`),
		},
		{
			name: "param tagid is missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"bidfloor": "0.04",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSovrn),
			},
			want: json.RawMessage(`{"bidfloor":0.04}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, string(openrtb_ext.BidderSovrn), nil)
			AssertJSON(t, tt.want, got)
		})
	}

}

func TestPrepareBidParamJSONForPartnerForConversant(t *testing.T) {

	type conversantTestObj struct {
		SiteID      string   `json:"site_id"`
		TagID       string   `json:"tag_id"`
		Secure      int      `json:"secure"`
		Position    int      `json:"position"`
		BidFloor    float64  `json:"bidfloor"`
		API         []int    `json:"api"`
		Protocols   []int    `json:"protocols"`
		Mimes       []string `json:"mimes"`
		Maxduration int      `json:"maxduration"`
	}

	mapping := map[string]interface{}{
		"site_id":     "12313",
		"tag_id":      "45343",
		"secure":      "1",
		"position":    "1",
		"bidfloor":    "0.12",
		"maxduration": "100",
		"mimes":       "[video/mp4,video/mp3]",
		"api":         "[1,2,3]",
		"protocols":   "[1]",
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250
	jsonString, _ := PrepareBidParamJSONForPartner(width, height, mapping, "adunit", string(openrtb_ext.BidderConversant), string(openrtb_ext.BidderConversant), nil)
	conv := new(conversantTestObj)
	if err := json.Unmarshal([]byte(jsonString), &conv); err != nil {
		t.Error("Incorrect Json Formed")
		return
	}

	expected := conversantTestObj{
		SiteID:      "12313",
		TagID:       "45343",
		Secure:      1,
		Position:    1,
		BidFloor:    0.12,
		Maxduration: 100,
		Mimes:       []string{"video/mp4", "video/mp3"},
		API:         []int{1, 2, 3},
		Protocols:   []int{1},
	}

	AssertJSON(t, json.RawMessage(GetJSON(expected)), jsonString)
}

func TestPrepareBidParamJSONForPartnerForRubicon(t *testing.T) {

	type videoObj struct {
		PlayerHeight int    `json:"playerHeight,omitempty"`
		PlayerWidth  int    `json:"playerWidth,omitempty"`
		SizeID       int    `json:"size_id,omitempty"`
		Language     string `json:"language,omitempty"`
	}

	type rubiconTestObj struct {
		AccountID int      `json:"accountId,omitempty"`
		ZoneID    int      `json:"zoneId,omitempty"`
		SiteID    int      `json:"siteId,omitempty"`
		Video     videoObj `json:"video,omitempty"`
	}

	vMap := map[string]interface{}{
		"playerWidth":  "1000",
		"playerHeight": "1000",
		"size_id":      "10",
		"language":     "eng",
	}

	mapping := map[string]interface{}{
		"accountId": "12313",
		"zoneId":    "45343",
		"siteId":    "12345",
		"video":     vMap,
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250
	jsonString, _ := PrepareBidParamJSONForPartner(width, height, mapping, "adunit", string(openrtb_ext.BidderRubicon), string(openrtb_ext.BidderRubicon), nil)
	actualR := new(rubiconTestObj)
	if err := json.Unmarshal([]byte(jsonString), &actualR); err != nil {
		t.Error("Incorrect Json Formed")
		return
	}

	expected := rubiconTestObj{
		AccountID: 12313,
		ZoneID:    45343,
		SiteID:    12345,
		Video: videoObj{
			PlayerHeight: 1000,
			PlayerWidth:  1000,
			SizeID:       10,
			Language:     "eng",
		},
	}

	AssertJSON(t, json.RawMessage(GetJSON(expected)), jsonString)
}

func TestPrepareBidParamJSONForPartnerForIndexExchange(t *testing.T) {

	type ixTestObj struct {
		SiteID string    `json:"siteId"`
		Size   [2]uint64 `json:"size"`
	}
	mapping := map[string]interface{}{
		"siteID": "12313",
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250
	jsonString, _ := PrepareBidParamJSONForPartner(width, height, mapping, "adunit", string(openrtb_ext.BidderIx), string(openrtb_ext.BidderIx), nil)
	ix := new(ixTestObj)
	if err := json.Unmarshal([]byte(jsonString), &ix); err != nil {
		t.Error("Incorrect Json Formed")
		return
	}

	if ix.SiteID != "12313" {
		t.Error("Incorrect value set for site_id")
	}

	if ix.Size[0] != 300 {
		t.Error("Incorrect value set for size[0]")
	}

	if ix.Size[1] != 250 {
		t.Error("Incorrect value set for size[1]")
	}
}

func TestPrepareBidParamJSONForPartnerForIndexExchangeForMissingSiteID(t *testing.T) {

	type ixTestObj struct {
		SiteID string    `json:"siteId"`
		Size   [2]uint64 `json:"size"`
	}
	mapping := map[string]interface{}{
		"memberId": "12313",
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250
	jsonString, _ := PrepareBidParamJSONForPartner(width, height, mapping, "adunit", string(openrtb_ext.BidderIx), string(openrtb_ext.BidderIx), nil)

	if jsonString != nil {
		t.Error("Value should not be set when siteID is missing")
	}
}

func TestPrepareBidParamJSONForPartnerForShareThrough(t *testing.T) {
	type shareThroughTestObj struct {
		Pkey string   `json:"pkey,omitempty"`
		Badv []string `json:"badv,omitempty"`
		Bcat []string `json:"bcat,omitempty"`
	}

	fieldMap := map[string]interface{}{
		"pkey": "12",
		"badv": []string{"1", "2"},
		"bcat": []string{"3", "4"},
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	jsonStrBuf, _ := PrepareBidParamJSONForPartner(width, height, fieldMap, "adunit", string(openrtb_ext.BidderSharethrough), string(openrtb_ext.BidderSharethrough), nil)

	var obj shareThroughTestObj
	if err := json.Unmarshal([]byte(jsonStrBuf), &obj); err != nil {
		t.Error("Failed to form json")
		return
	}

	if obj.Pkey != fieldMap["pkey"] {
		t.Error("wrong pkey value set")
		return
	}

	assert.EqualValues(t, fieldMap["badv"].([]string), obj.Badv)
	assert.EqualValues(t, fieldMap["bcat"].([]string), obj.Bcat)
}

func TestPrepareBidParamJSONForPartnerForShareThroughWhenPkeyMissing(t *testing.T) {
	type shareThroughTestObj struct {
		Pkey       string `json:"pkey,omitempty"`
		Iframe     bool   `json:"iframe,omitempty"`
		IframeSize [2]int `json:"iframeSize,omitempty"`
	}

	fieldMap := map[string]interface{}{
		"iframe":     "true",
		"iframeSize": "[300, 250]",
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	jsonStrBuf, _ := PrepareBidParamJSONForPartner(width, height, fieldMap, "adunit", string(openrtb_ext.BidderSharethrough), string(openrtb_ext.BidderSharethrough), nil)

	if jsonStrBuf != nil {
		t.Error("JSON should be empty string")
		return
	}
}

func TestPrepareBidParamJSONForPartnerForTripleLift(t *testing.T) {
	type tripleLiftTestObj struct {
		InventoryCode string  `json:"inventoryCode,omitempty"`
		Floor         float64 `json:"floor,omitempty"`
	}

	fieldMap := map[string]interface{}{
		"inventoryCode": "121",
		"floor":         "9.11",
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	jsonStrBuf, _ := PrepareBidParamJSONForPartner(width, height, fieldMap, "adunit", string(openrtb_ext.BidderTriplelift), string(openrtb_ext.BidderTriplelift), nil)
	var obj tripleLiftTestObj
	if err := json.Unmarshal([]byte(jsonStrBuf), &obj); err != nil {
		t.Error("Failed to form json")
		return
	}

	if obj.InventoryCode != "121" {
		t.Error("wrong inventoryCode value set")
		return
	}

	floor, _ := strconv.ParseFloat(fmt.Sprintf("%v", fieldMap["floor"]), 64)
	if floor != obj.Floor {
		t.Error("wrong floor value set")
		return
	}
}

func TestPrepareBidParamJSONForPartnerForTripleLiftWhenFloorMissing(t *testing.T) {
	type tripleLiftTestObj struct {
		InventoryCode string  `json:"inventoryCode,omitempty"`
		Floor         float64 `json:"floor,omitempty"`
	}

	fieldMap := map[string]interface{}{
		"inventoryCode": "121",
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	jsonStrBuf, _ := PrepareBidParamJSONForPartner(width, height, fieldMap, "adunit", string(openrtb_ext.BidderTriplelift), string(openrtb_ext.BidderTriplelift), nil)
	var obj tripleLiftTestObj
	if err := json.Unmarshal([]byte(jsonStrBuf), &obj); err != nil {
		t.Error("Failed to form json")
		return
	}

	if obj.Floor != 0 {
		t.Error("floor should be 0")
		return
	}
}

func TestPrepareBidParamJSONForPartnerForTripleLiftWhenRequiredParamMissing(t *testing.T) {
	type tripleLiftTestObj struct {
		InventoryCode string  `json:"inventoryCode,omitempty"`
		Floor         float64 `json:"floor,omitempty"`
	}

	fieldMap := map[string]interface{}{
		"floor": "9.11",
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	jsonStrBuf, _ := PrepareBidParamJSONForPartner(width, height, fieldMap, "adunit", string(openrtb_ext.BidderTriplelift), string(openrtb_ext.BidderTriplelift), nil)
	if jsonStrBuf != nil {
		t.Error("JSON should be empty")
		return
	}
}

func TestPrepareBidParamJSONForPartnerForImproveDigitalWithPlacementIdAndPublisherId(t *testing.T) {
	type improveDigitalTestObj struct {
		PlacementID  int    `json:"placementId,omitempty"`
		PublisherID  int    `json:"publisherId,omitempty"`
		PlacementKey string `json:"placementKey,omitempty"`
	}

	fieldMap := map[string]interface{}{
		"placementId": "121",
		"publisherId": "911",
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	jsonStrBuf, _ := PrepareBidParamJSONForPartner(width, height, fieldMap, "adunit", string(openrtb_ext.BidderImprovedigital), string(openrtb_ext.BidderImprovedigital), nil)
	var obj improveDigitalTestObj
	if err := json.Unmarshal([]byte(jsonStrBuf), &obj); err != nil {
		t.Error("Failed to form json")
		return
	}

	if obj.PlacementID != 121 {
		t.Error("wrong placementId value set")
		return
	}

	if obj.PublisherID != 911 {
		t.Error("wrong publisherId value set")
		return
	}

	if obj.PlacementKey != "" {
		t.Error("wrong placementKey value set")
		return
	}
}

func TestPrepareBidParamJSONForPartnerTelaria(t *testing.T) {
	type telariaTestObj struct {
		SeatCode string `json:"seatCode"`
		AdCode   string `json:"adCode,omitempty"`
	}

	fieldMap := map[string]interface{}{
		"seatCode": "test_seat_code",
		"adCode":   "test_ad_code",
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	jsonStrBuf, _ := PrepareBidParamJSONForPartner(width, height, fieldMap, "adunit", string(openrtb_ext.BidderTelaria), string(openrtb_ext.BidderTelaria), nil)
	var obj telariaTestObj
	if err := json.Unmarshal([]byte(jsonStrBuf), &obj); err != nil {
		t.Errorf("Failed to form json: %v", err)
		return
	}

	if obj.AdCode != "test_ad_code" {
		t.Error("wrong adCode value set")
		return
	}

	if obj.SeatCode != "test_seat_code" {
		t.Error("wrong seatCode value set")
		return
	}
}

func TestPrepareBidParamJSONForPartnerTelariaWithoutSeatCode(t *testing.T) {
	type telariaTestObj struct {
		SeatCode string `json:"seatCode"`
		AdCode   string `json:"adCode,omitempty"`
	}

	fieldMap := map[string]interface{}{
		"adCode": "test_ad_code",
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	jsonStrBuf, _ := PrepareBidParamJSONForPartner(width, height, fieldMap, "adunit", string(openrtb_ext.BidderTelaria), string(openrtb_ext.BidderTelaria), nil)
	if jsonStrBuf != nil {
		t.Error("JSON should be empty")
		return
	}
}

func TestPrepareBidParamJSONForPartnerForImproveDigitalWithoutPlacementId(t *testing.T) {
	type improveDigitalTestObj struct {
		PlacementID  int    `json:"placementId,omitempty"`
		PublisherID  int    `json:"publisherId,omitempty"`
		PlacementKey string `json:"placementKey,omitempty"`
	}

	fieldMap := map[string]interface{}{
		"publisherId": "911",
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	jsonStrBuf, _ := PrepareBidParamJSONForPartner(width, height, fieldMap, "adunit", string(openrtb_ext.BidderImprovedigital), string(openrtb_ext.BidderImprovedigital), nil)
	if jsonStrBuf != nil {
		t.Error("jsonStrBuf should be nil")
		return
	}
}

func TestPrepareBidParamJSONForPartnerSpotx(t *testing.T) {
	type spotxTestObject struct {
		ChannelID  string  `json:"channel_id"`
		AdUnit     string  `json:"ad_unit"`
		Secure     bool    `json:"secure,omitempty"`
		AdVolume   float64 `json:"ad_volume,omitempty"`
		PriceFloor int     `json:"price_floor,omitempty"`
		HideSkin   bool    `json:"hide_skin,omitempty"`
	}

	fieldMap := map[string]interface{}{
		"channel_id":  "12345",
		"ad_unit":     "outstream",
		"secure":      true,
		"ad_volume":   19.1,
		"price_floor": 12.0,
		"hide_skin":   false,
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	jsonStrBuf, _ := PrepareBidParamJSONForPartner(width, height, fieldMap, "adunit", string(openrtb_ext.BidderSpotX), string(openrtb_ext.BidderSpotX), nil)
	var obj spotxTestObject
	if err := json.Unmarshal([]byte(jsonStrBuf), &obj); err != nil {
		t.Errorf("Failed to form json: %v", err)
		return
	}

	if obj.ChannelID != "12345" {
		t.Error("wrong channel_id value set")
		return
	}

	if obj.AdUnit != "outstream" {
		t.Error("wrong ad_unit value set")
		return
	}

	if !obj.Secure {
		t.Error("wrong secure value set")
		return
	}

	if obj.AdVolume != 19.1 {
		t.Error("wrong ad_volume value set")
		return
	}

	if obj.PriceFloor != 12 {
		t.Error("wrong price_floor value set")
		return
	}

	if obj.HideSkin {
		t.Error("wrong secure value set")
		return
	}
}

func TestPrepareBidParamJSONForPartnerSpotWithoutChannelId(t *testing.T) {
	type spotxTestObject struct {
		ChannelID  string  `json:"channel_id"`
		AdUnit     string  `json:"ad_unit"`
		Secure     bool    `json:"secure,omitempty"`
		AdVolume   float64 `json:"ad_volume,omitempty"`
		PriceFloor int     `json:"price_floor,omitempty"`
		HideSkin   bool    `json:"hide_skin,omitempty"`
	}

	fieldMap := map[string]interface{}{
		"ad_unit":     "outstream",
		"secure":      true,
		"ad_volume":   19.1,
		"price_floor": 12.0,
		"hide_skin":   false,
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	jsonStrBuf, _ := PrepareBidParamJSONForPartner(width, height, fieldMap, "adunit", string(openrtb_ext.BidderSpotX), string(openrtb_ext.BidderSpotX), nil)
	if jsonStrBuf != nil {
		t.Error("JSON should be empty string")
		return
	}
}

func TestPrepareBidParamJSONForPartnerOpenX(t *testing.T) {

	type OpenXTestObj struct {
		DelDomain string `json:"delDomain"`
		Unit      string `json:"unit"`
	}

	mapping := map[string]interface{}{
		"delDomain": "test.openx.domain",
		"unit":      "45343",
	}
	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	jsonString, _ := PrepareBidParamJSONForPartner(width, height, mapping, "adunit", string(openrtb_ext.BidderOpenx), string(openrtb_ext.BidderOpenx), nil)
	openxObj := new(OpenXTestObj)
	if err := json.Unmarshal([]byte(jsonString), &openxObj); err != nil {
		t.Error("Incorrect Json Formed")
		return
	}

	if openxObj.DelDomain != "test.openx.domain" {
		t.Error("Incorrect value set for delDomain")
	}

	if openxObj.Unit != "45343" {
		t.Error("Incorrect value set for unit")
	}
}

func TestPrepareBidParamJSONForPartnerOpenXMissingParameters(t *testing.T) {

	type OpenXTestObj struct {
		DelDomain string `json:"delDomain"`
		Unit      string `json:"unit"`
	}

	mapping := map[string]interface{}{}
	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	jsonString, _ := PrepareBidParamJSONForPartner(width, height, mapping, "adunit", string(openrtb_ext.BidderOpenx), string(openrtb_ext.BidderOpenx), nil)
	openxObj := new(OpenXTestObj)
	if err := json.Unmarshal([]byte(jsonString), &openxObj); err != nil {
		t.Error("Incorrect Json Formed")
		return
	}

	if openxObj.DelDomain != "" || openxObj.Unit != "" {
		t.Error("Incorrect value set for delDomain or unit")
	}
}

func TestPrepareBidParamJSONForPartnerSynacorMedia(t *testing.T) {
	type args struct {
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All params present",
			args: args{
				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"seatId": "testSeatId",
					"tagId":  "testTagId",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderImds),
				bidderCode:  string(openrtb_ext.BidderImds),
			},
			want: json.RawMessage(`{"seatId":"testSeatId","tagId":"testTagId"}`),
		},
		{
			name: "required param seatId missing",
			args: args{
				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"tagId": "testTagId",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderImds),
				bidderCode:  string(openrtb_ext.BidderImds),
			},
			want: nil,
		},
		{
			name: "param tagId missing",
			args: args{
				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"seatId": "testSeatId",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderImds),
				bidderCode:  string(openrtb_ext.BidderImds),
			},
			want: json.RawMessage(`{"seatId":"testSeatId"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerGumGum(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All params present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"zone": "testZone",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderGumGum),
				bidderCode:  string(openrtb_ext.BidderGumGum),
			},
			want: json.RawMessage(`{"zone":"testZone"}`),
		},
		{
			name: "required params are missing",
			args: args{
				reqID:       "",
				width:       nil,
				height:      nil,
				fieldMap:    nil,
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderGumGum),
				bidderCode:  string(openrtb_ext.BidderGumGum),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerYieldone(t *testing.T) {
	type args struct {
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All params present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"placementId": "testplacementId",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderYieldone),
				bidderCode:  string(openrtb_ext.BidderYieldone),
			},
			want: json.RawMessage(`{"placementId":"testplacementId"}`),
		},
		{
			name: "required param id is missing",
			args: args{
				width:       nil,
				height:      nil,
				fieldMap:    map[string]interface{}{},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderYieldone),
				bidderCode:  string(openrtb_ext.BidderYieldone),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerDistrictmDMX(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All required params present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"memberid": "testmemberid",
					"dmxid":    "testdmxid",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderDmx),
				bidderCode:  string(openrtb_ext.BidderDmx),
			},
			want: json.RawMessage(`{"dmxid":"testdmxid","memberid":"testmemberid","tagid":"testdmxid"}`),
		},
		{
			name: "memberid is missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"dmxid": "testdmxid",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderDmx),
				bidderCode:  string(openrtb_ext.BidderDmx),
			},
			want: nil,
		},
		{
			name: "dmxid is missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"memberid": "testmemberid",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderDmx),
				bidderCode:  string(openrtb_ext.BidderDmx),
			},
			want: json.RawMessage(`{"memberid":"testmemberid"}`),
		},
		{
			name: "All params are missing",
			args: args{
				reqID:       "",
				width:       nil,
				height:      nil,
				fieldMap:    nil,
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderDmx),
				bidderCode:  string(openrtb_ext.BidderDmx),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerAdGenration(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All params present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"id": "testid",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderAdgeneration),
				bidderCode:  string(openrtb_ext.BidderAdgeneration),
			},
			want: json.RawMessage(`{"id":"testid"}`),
		},
		{
			name: "required param id is missing",
			args: args{
				reqID:       "",
				width:       nil,
				height:      nil,
				fieldMap:    map[string]interface{}{},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderAdgeneration),
				bidderCode:  string(openrtb_ext.BidderAdgeneration),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerBeachfront(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All params present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"appId":    "testAppID",
					"bidfloor": "0.1",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderBeachfront),
				bidderCode:  string(openrtb_ext.BidderBeachfront),
			},
			want: json.RawMessage(`{"appId":"testAppID","bidfloor":0.1,"videoResponseType":"adm"}`),
		},
		{
			name: "appId missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"bidfloor": "0.1",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderBeachfront),
				bidderCode:  string(openrtb_ext.BidderBeachfront),
			},
			want: nil,
		},
		{
			name: "bidfloor missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"bidfloor": "0.1",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderBeachfront),
				bidderCode:  string(openrtb_ext.BidderBeachfront),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, json.RawMessage(tt.want), json.RawMessage(got))
		})
	}
}

func TestPrepareBidParamJSONForPartnerVRTCAL(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "Dummy param present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"dummyParam": "2",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderVrtcal),
				bidderCode:  string(openrtb_ext.BidderVrtcal),
			},
			want: json.RawMessage(`{"just_an_unused_vrtcal_param":"2"}`),
		},
		{
			name: "Dummy param missing",
			args: args{
				reqID:       "",
				width:       nil,
				height:      nil,
				fieldMap:    map[string]interface{}{},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderVrtcal),
				bidderCode:  string(openrtb_ext.BidderVrtcal),
			},
			want: json.RawMessage(`{"just_an_unused_vrtcal_param":"1"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerInMobi(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All params present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"plc": "1234",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderInMobi),
				bidderCode:  string(openrtb_ext.BidderInMobi),
			},
			want: json.RawMessage(`{"plc":"1234"}`),
		},
		{
			name: "required param plc is missing",
			args: args{
				reqID:       "",
				width:       nil,
				height:      nil,
				fieldMap:    map[string]interface{}{},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderInMobi),
				bidderCode:  string(openrtb_ext.BidderInMobi),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerTappx(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All params present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"tappxkey": "key1",
					"endpoint": "endpoint1",
					"bidfloor": "0.2",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderTappx),
				bidderCode:  string(openrtb_ext.BidderTappx),
			},
			want: json.RawMessage(`{"tappxkey":"key1","endpoint":"endpoint1","bidfloor":0.2}`),
		},
		{
			name: "tappxkey missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"endpoint": "endpoint1",
					"bidfloor": "0.2",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderTappx),
				bidderCode:  string(openrtb_ext.BidderTappx),
			},
			want: nil,
		},
		{
			name: "endpoint missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"tappxkey": "key1",
					"bidfloor": "0.2",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderTappx),
				bidderCode:  string(openrtb_ext.BidderTappx),
			},
			want: nil,
		},
		{
			name: "bidfloor missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"tappxkey": "key1",
					"endpoint": "endpoint1",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderTappx),
				bidderCode:  string(openrtb_ext.BidderTappx),
			},
			want: json.RawMessage(`{"tappxkey":"key1","endpoint":"endpoint1"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerNobid(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All params present with correct mapping value datatype",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"siteId":      1234,
					"placementId": 5678,
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderNoBid),
			},
			want: json.RawMessage(`{"siteId":1234,"placementId":5678}`),
		},
		{
			name: "All params present with string mapping value datatype",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"siteId":      "1234",
					"placementId": "5678",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderNoBid),
				bidderCode:  string(openrtb_ext.BidderNoBid),
			},
			want: json.RawMessage(`{"siteId":1234,"placementId":5678}`),
		},
		{
			name: "required param siteId is missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"placementId": "5678",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderNoBid),
				bidderCode:  string(openrtb_ext.BidderNoBid),
			},
			want: nil,
		},
		{
			name: "required param siteId is not an integer",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"siteId": "abc",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderNoBid),
				bidderCode:  string(openrtb_ext.BidderNoBid),
			},
			want: nil,
		},
		{
			name: "optional param placementId is missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"siteId": "1234",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderNoBid),
				bidderCode:  string(openrtb_ext.BidderNoBid),
			},
			want: json.RawMessage(`{"siteId":1234}`),
		},
		{
			name: "optional param placementId is not an integer",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"siteId":      "1234",
					"placementId": "abc",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderNoBid),
				bidderCode:  string(openrtb_ext.BidderNoBid),
			},
			want: json.RawMessage(`{"siteId":1234}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerAudienceNetwork(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All params present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"placementId": "testPlacementId",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderAudienceNetwork),
				bidderCode:  string(openrtb_ext.BidderAudienceNetwork),
			},
			want: json.RawMessage(`{"placementId":"testPlacementId"}`),
		},
		{
			name: "required param id is missing",
			args: args{
				reqID:       "",
				width:       nil,
				height:      nil,
				fieldMap:    map[string]interface{}{},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderAudienceNetwork),
				bidderCode:  string(openrtb_ext.BidderAudienceNetwork),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerGrid(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All params present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"uid": "1234",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderGrid),
				bidderCode:  string(openrtb_ext.BidderGrid),
			},
			want: json.RawMessage(`{"uid":1234}`),
		},
		{
			name: "required param uid is missing",
			args: args{
				reqID:       "",
				width:       nil,
				height:      nil,
				fieldMap:    map[string]interface{}{},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderGrid),
				bidderCode:  string(openrtb_ext.BidderGrid),
			},
			want: nil,
		},
		{
			name: "required param uid is not an integer",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"uid": "abc",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderGrid),
				bidderCode:  string(openrtb_ext.BidderGrid),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerSmartAdServer(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "Network ID present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"networkId": "1234",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmartAdserver),
				bidderCode:  string(openrtb_ext.BidderSmartAdserver),
			},
			want: json.RawMessage(`{"networkId":1234}`),
		},
		{
			name: "Network ID missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"siteId": "1234",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmartAdserver),
				bidderCode:  string(openrtb_ext.BidderSmartAdserver),
			},
			want: nil,
		},
		{
			name: "Network ID not an integer",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"networkId": "test",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmartAdserver),
				bidderCode:  string(openrtb_ext.BidderSmartAdserver),
			},
			want: nil,
		},
		{
			name: "All params present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"networkId": "1234",
					"siteId":    "3456",
					"pageId":    "8901",
					"formatId":  "7612",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmartAdserver),
				bidderCode:  string(openrtb_ext.BidderSmartAdserver),
			},
			want: json.RawMessage(`{"networkId":1234,"siteId":3456,"pageId":8901,"formatId":7612}`),
		},
		{
			name: "Site ID and Format ID present, page ID missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"networkId": "1234",
					"siteId":    "3456",
					"formatId":  "7612",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmartAdserver),
				bidderCode:  string(openrtb_ext.BidderSmartAdserver),
			},
			want: json.RawMessage(`{"networkId":1234}`),
		},
		{
			name: "Site ID and Page ID present, Format ID missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"networkId": "1234",
					"siteId":    "3456",
					"pageId":    "7612",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmartAdserver),
				bidderCode:  string(openrtb_ext.BidderSmartAdserver),
			},
			want: json.RawMessage(`{"networkId":1234}`),
		},
		{
			name: "Format ID and Page ID present, Site ID missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"networkId": "1234",
					"formatId":  "3456",
					"pageId":    "7612",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmartAdserver),
				bidderCode:  string(openrtb_ext.BidderSmartAdserver),
			},
			want: json.RawMessage(`{"networkId":1234}`),
		},
		{
			name: "Site ID and Format ID present, page ID not an integer",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"networkId": "1234",
					"siteId":    "3456",
					"formatId":  "7612",
					"pageId":    "test",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmartAdserver),
				bidderCode:  string(openrtb_ext.BidderSmartAdserver),
			},
			want: json.RawMessage(`{"networkId":1234}`),
		},
		{
			name: "Site ID and Page ID present, Format ID not an integer",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"networkId": "1234",
					"siteId":    "3456",
					"pageId":    "7612",
					"formatId":  "test",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmartAdserver),
				bidderCode:  string(openrtb_ext.BidderSmartAdserver),
			},
			want: json.RawMessage(`{"networkId":1234}`),
		},
		{
			name: "Format ID and Page ID present, Site ID not an integer",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"networkId": "1234",
					"formatId":  "3456",
					"pageId":    "7612",
					"siteId":    "test",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmartAdserver),
				bidderCode:  string(openrtb_ext.BidderSmartAdserver),
			},
			want: json.RawMessage(`{"networkId":1234}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerSmaato(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "all_missing",
			args: args{
				reqID:       "",
				width:       nil,
				height:      nil,
				fieldMap:    map[string]interface{}{},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmaato),
				bidderCode:  string(openrtb_ext.BidderSmaato),
			},
			want: nil,
		},
		{
			name: "publisherId missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"adspaceId": "1234",
					"adbreakId": "4567",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmaato),
				bidderCode:  string(openrtb_ext.BidderSmaato),
			},
			want: nil,
		},
		{
			name: "adspaceId missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"publisherId": "1234",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmaato),
				bidderCode:  string(openrtb_ext.BidderSmaato),
			},
			want: json.RawMessage(`{"publisherId": "1234"}`),
		},
		{
			name: "adbreakId missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"publisherId": "1234",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmaato),
				bidderCode:  string(openrtb_ext.BidderSmaato),
			},
			want: json.RawMessage(`{"publisherId": "1234"}`),
		},
		{
			name: "publisherId_adspaceId__present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"publisherId": "1234",
					"adspaceId":   "3456",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmaato),
				bidderCode:  string(openrtb_ext.BidderSmaato),
			},
			want: json.RawMessage(`{"publisherId": "1234","adspaceId": "3456"}`),
		},
		{
			name: "publisherId_adbreakId__present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"publisherId": "1234",
					"adbreakId":   "3456",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmaato),
				bidderCode:  string(openrtb_ext.BidderSmaato),
			},
			want: json.RawMessage(`{"publisherId": "1234","adbreakId": "3456"}`),
		},
		{
			name: "adspaceId_and_adbreakId_present",
			args: args{
				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"adspaceId": "1234",
					"adbreakId": "7899",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmaato),
				bidderCode:  string(openrtb_ext.BidderSmaato),
			},
			want: nil,
		},
		{
			name: "all_params_are_present",
			args: args{
				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"publisherId": "1234",
					"adspaceId":   "3456",
					"adbreakId":   "7899",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSmaato),
				bidderCode:  string(openrtb_ext.BidderSmaato),
			},
			want: json.RawMessage(`{"publisherId": "1234","adspaceId": "3456","adbreakId": "7899"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerPangle(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "Required param present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"token": "testToken",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderPangle),
				bidderCode:  string(openrtb_ext.BidderPangle),
			},
			want: json.RawMessage(`{"token":"testToken"}`),
		},
		{
			name: "required params are missing",
			args: args{
				reqID:       "",
				width:       nil,
				height:      nil,
				fieldMap:    nil,
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderPangle),
				bidderCode:  string(openrtb_ext.BidderPangle),
			},
			want: nil,
		},
		{
			name: "dependent param 'placementid' is missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"token": "testToken",
					"appid": "testAppID",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderPangle),
				bidderCode:  string(openrtb_ext.BidderPangle),
			},
			want: nil,
		},
		{
			name: "dependent param 'appid' is missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"token":       "testToken",
					"placementid": "testPlacementID",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderPangle),
				bidderCode:  string(openrtb_ext.BidderPangle),
			},
			want: nil,
		},
		{
			name: "all params are present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"token":       "testToken",
					"placementid": "testPlacementID",
					"appid":       "testAppID",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderPangle),
				bidderCode:  string(openrtb_ext.BidderPangle),
			},
			want: json.RawMessage(`{"token":"testToken","placementid":"testPlacementID","appid":"testAppID"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerSonobi(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "ad_unit missing but placement_id present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"placement_id": "testPlacementId",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSonobi),
			},
			want: json.RawMessage(`{"TagID":"testPlacementId"}`),
		},
		{
			name: "placement_id missing but ad_unit present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"ad_unit": "testAdUnit",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSonobi),
			},
			want: json.RawMessage(`{"TagID":"testAdUnit"}`),
		},
		{
			name: "empty fieldmap",
			args: args{
				reqID:       "",
				width:       nil,
				height:      nil,
				fieldMap:    map[string]interface{}{},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSonobi),
			},
			want: nil,
		},
		{
			name: "ad_unit & placement_id both are present",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"ad_unit":      "testAdUnit",
					"placement_id": "testPlacementId",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderSonobi),
			},
			want: json.RawMessage(`{"TagID":"testAdUnit"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, string(openrtb_ext.BidderSonobi), nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerImproveDigital(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "All params present",
			args: args{

				width:  ptrutil.ToPtr[int64](300),
				height: ptrutil.ToPtr[int64](250),
				fieldMap: map[string]interface{}{
					"placementId":  "1234",
					"publisherId":  "5678",
					"placementKey": "key1",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderImprovedigital),
				bidderCode:  string(openrtb_ext.BidderImprovedigital),
			},
			want: json.RawMessage(`{"placementId":1234,"publisherId":5678,"size":{"w":300,"h":250}}`),
		},
		{
			name: "PlacementId present, publisherId and placementKey missing",
			args: args{

				width:  ptrutil.ToPtr[int64](300),
				height: ptrutil.ToPtr[int64](250),
				fieldMap: map[string]interface{}{
					"placementId": "1234",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderImprovedigital),
				bidderCode:  string(openrtb_ext.BidderImprovedigital),
			},
			want: nil,
		},
		{
			name: "PlacementId and publisherId present, placementKey missing",
			args: args{

				width:  ptrutil.ToPtr[int64](300),
				height: ptrutil.ToPtr[int64](250),
				fieldMap: map[string]interface{}{
					"placementId": "1234",
					"publisherId": "5678",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderImprovedigital),
				bidderCode:  string(openrtb_ext.BidderImprovedigital),
			},
			want: json.RawMessage(`{"placementId":1234,"publisherId":5678,"size":{"w":300,"h":250}}`),
		},
		{
			name: "PlacementId absent, publisherId and placementKey present",
			args: args{

				width:  ptrutil.ToPtr[int64](300),
				height: ptrutil.ToPtr[int64](250),
				fieldMap: map[string]interface{}{
					"publisherId":  "5678",
					"placementKey": "key1",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderImprovedigital),
				bidderCode:  string(openrtb_ext.BidderImprovedigital),
			},
			want: nil,
		},
		{
			name: "required params missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"placementKey": "key1",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderImprovedigital),
				bidderCode:  string(openrtb_ext.BidderImprovedigital),
			},
			want: nil,
		},
		{
			name: "required param placementId is not an integer",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"placementId":  "abc",
					"placementKey": "key1",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderImprovedigital),
				bidderCode:  string(openrtb_ext.BidderImprovedigital),
			},
			want: nil,
		},
		{
			name: "required param publisherId is not an integer",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"publisherId":  "abc",
					"placementKey": "key1",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderImprovedigital),
				bidderCode:  string(openrtb_ext.BidderImprovedigital),
			},
			want: nil,
		},
		{
			name: "optional param size is missing",
			args: args{

				width:  nil,
				height: nil,
				fieldMap: map[string]interface{}{
					"placementId": "1234",
					"publisherId": "5678",
				},
				slotKey:     "",
				adapterName: string(openrtb_ext.BidderImprovedigital),
				bidderCode:  string(openrtb_ext.BidderImprovedigital),
			},
			want: json.RawMessage(`{"placementId":1234, "publisherId":5678}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerOutbrain(t *testing.T) {
	type args struct {
		reqID       string
		width       *int64
		height      *int64
		fieldMap    map[string]interface{}
		slotKey     string
		adapterName string
		bidderCode  string
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "id within publisher object present",
			args: args{
				fieldMap: map[string]interface{}{
					"publisher": map[string]interface{}{
						"id": "myid",
					},
				},
				adapterName: string(openrtb_ext.BidderOutbrain),
			},
			want: json.RawMessage(`{"publisher":{"id":"myid"}}`),
		},
		{
			name: "Empty publisher object",
			args: args{
				fieldMap: map[string]interface{}{
					"publisher": map[string]interface{}{},
				},
				adapterName: string(openrtb_ext.BidderOutbrain),
			},
			want: nil,
		},
		{
			name: "NO publisher object",
			args: args{
				fieldMap: map[string]interface{}{
					"PubLisheR": map[string]interface{}{
						"ID": "myid",
					},
				},
				adapterName: string(openrtb_ext.BidderOutbrain),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := PrepareBidParamJSONForPartner(tt.args.width, tt.args.height, tt.args.fieldMap, tt.args.slotKey, tt.args.adapterName, tt.args.bidderCode, nil)
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestPrepareBidParamJSONForPartnerForAppnexus(t *testing.T) {

	mapping := map[string]interface{}{
		"invCode":        "abc",
		"placementId":    "4978056",
		"query":          "def",
		"reserve":        "1",
		"usePaymentRule": "true",
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	impExt := new(models.ImpExtension)
	impExt.Bidder = formBidderKeywordsMap()

	jsonString, _ := PrepareBidParamJSONForPartner(width, height, mapping, "adunit", string(openrtb_ext.BidderAppnexus), string(openrtb_ext.BidderAppnexus), impExt)
	actualR := new(openrtb_ext.ExtImpAppnexus)
	if err := json.Unmarshal([]byte(jsonString), &actualR); err != nil {

		t.Error("Incorrect Json Formed: ERR: ", err)
		t.Error("Incorrect Json Formed: ", jsonString)
		return
	}

	upr := new(bool)
	*upr = true
	var kw openrtb_ext.ExtImpAppnexusKeywords = "key1=val1,key1=val2,key2=val3,key2=val4"

	expected := openrtb_ext.ExtImpAppnexus{
		DeprecatedPlacementId: 4978056,
		LegacyInvCode:         "abc",
		Reserve:               1.0,
		UsePaymentRule:        upr,
		Keywords:              kw,
	}

	AssertJSON(t, json.RawMessage(GetJSON(expected)), json.RawMessage(GetJSON(actualR)))
}

func TestPrepareBidParamJSONForPartnerForAppnexusAlias(t *testing.T) {

	mapping := map[string]interface{}{
		"invCode":        "abc",
		"placementId":    "4978056",
		"query":          "def",
		"reserve":        "1",
		"usePaymentRule": "true",
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	impExt := new(models.ImpExtension)
	impExt.Bidder = formBidderKeywordsMapForAppnexusAlias()

	jsonString, _ := PrepareBidParamJSONForPartner(width, height, mapping, "adunit", string(openrtb_ext.BidderAppnexus), "appnexus-alias", impExt)
	actualR := new(openrtb_ext.ExtImpAppnexus)
	if err := json.Unmarshal([]byte(jsonString), &actualR); err != nil {

		t.Error("Incorrect Json Formed: ERR: ", err)
		t.Error("Incorrect Json Formed: ", jsonString)
		return
	}

	upr := new(bool)
	*upr = true
	var kw openrtb_ext.ExtImpAppnexusKeywords = "key1=val1,key1=val2,key2=val3,key2=val4"

	expected := openrtb_ext.ExtImpAppnexus{
		DeprecatedPlacementId: 4978056,
		LegacyInvCode:         "abc",
		Reserve:               1.0,
		UsePaymentRule:        upr,
		Keywords:              kw,
	}

	AssertJSON(t, json.RawMessage(GetJSON(expected)), json.RawMessage(GetJSON(actualR)))
}

func TestPrepareBidParamJSONForPartnerForAppnexusForIgnoredKeys(t *testing.T) {

	mapping := map[string]interface{}{
		"placementId": "4978056",
		"video": map[string]interface{}{
			"frameworks":      []int{0, 1, 2},
			"playback_method": []string{"auto_play_sound_on"},
			"skippable":       true,
		},
	}

	width := new(int64)
	*width = 300
	height := new(int64)
	*height = 250

	expectedJSONString := json.RawMessage(`{"placementId":4978056}`)
	actualJSONString, _ := PrepareBidParamJSONForPartner(width, height, mapping, "adunit", string(openrtb_ext.BidderAppnexus), string(openrtb_ext.BidderAppnexus), nil)
	AssertJSON(t, expectedJSONString, actualJSONString)
}

func Test_builderPubMatic(t *testing.T) {
	expected := JSONObject{"publisherId": "301", "pmzoneid": "Zone1", "dctr": "value1", "keywords": "test"}

	type args struct {
		params BidderParameters
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name:    "TEST1",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"publisherId": "301", "pmzoneid": "Zone1", "dctr": "value1", "keywords": "test"}}},
			want:    json.RawMessage(GetJSON(expected)),
			wantErr: false,
		},
		{
			name:    "When bidderparam BidviewabilityScore is present, read and pass it",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"publisherId": "301", "bidViewability": JSONObject{"createdAt": 1666155076240, "lastViewed": 3171.100000023842, "rendered": 131, "totalViewTime": 15468, "updatedAt": 1666296333802, "viewed": 80}}}},
			want:    json.RawMessage(`{"publisherId": "301", "bidViewability": {"createdAt": 1666155076240, "lastViewed": 3171.100000023842, "rendered": 131, "totalViewTime": 15468, "updatedAt": 1666296333802, "viewed": 80}}`),
			wantErr: false,
		},
		{
			name:    "When bidderparam BidviewabilityScore is present, but with limited fields ,read and pass it",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"publisherId": "301", "bidViewability": JSONObject{"createdAt": 1666155076240, "rendered": 131, "updatedAt": 1666296333802, "viewed": 0}}}},
			want:    json.RawMessage(`{"publisherId": "301", "bidViewability": {"createdAt": 1666155076240, "rendered": 131,"updatedAt": 1666296333802, "viewed": 0}}`),
			wantErr: false,
		},
		{
			name:    "When bidderparam BidviewabilityScore is present with invalid json fields,ignore passing bidviewability object",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"publisherId": "301", "bidViewability": json.RawMessage(`{"createdAt": 1666155076240, "rendered: 131, "updatedAt: 1666296333802, "viewed}": 0}`)}}},
			want:    json.RawMessage(`{"publisherId": "301"}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builderPubMatic(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("builderPubMatic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			AssertJSON(t, tt.want, got)

		})
	}
}

func Test_builderAppNexus(t *testing.T) {
	expected := JSONObject{"placementId": 0, "keywords": "test", "generate_ad_pod_id": true, "member": "958"}

	type args struct {
		params BidderParameters
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name:    "TEST1",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"keywords": "test", "generate_ad_pod_id": true, "member": "958"}}},
			want:    json.RawMessage(GetJSON(expected)),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builderAppNexus(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("builderAppNexus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			AssertJSON(t, tt.want, got)
		})
	}
}

func Test_builderPulsePoint(t *testing.T) {
	expected := JSONObject{"cp": 0, "ct": 0, "cf": "70"}
	type args struct {
		params BidderParameters
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name:    "TEST1",
			args:    args{params: BidderParameters{SlotKey: "adunit@70"}},
			want:    json.RawMessage(GetJSON(expected)),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builderPulsePoint(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("builderPulsePoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			AssertJSON(t, tt.want, got)
		})
	}
}

func Test_builderApacdex(t *testing.T) {
	type args struct {
		params BidderParameters
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name:    "Valid Scenerio (oneOf siteId or PlacementId) is present with FloorPrice and geo Object",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"siteId": "test123", "floorPrice": 0.999223, "geo": JSONObject{"lat": 17.98928, "lon": 99.7741712, "accuracy": 20}}}},
			want:    json.RawMessage(GetJSON(JSONObject{"siteId": "test123", "floorPrice": 0.999223, "geo": JSONObject{"lat": 17.98928, "lon": 99.7741712, "accuracy": 20}})),
			wantErr: false,
		},
		{
			name:    "Invalid Scenerio both siteId and PlacementId are present with FloorPrice, select siteId  & ignore placementId",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"siteId": "test123", "placementId": "testPlacementid", "floorPrice": 0.999223}}},
			want:    json.RawMessage(GetJSON(JSONObject{"siteId": "test123", "floorPrice": 0.999223})),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf siteId or PlacementId) is present and FloorPrice is absent, ignore FloorPrice param",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"siteId": "test123"}}},
			want:    json.RawMessage(GetJSON(JSONObject{"siteId": "test123"})),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio only PlacementId is present, expect only placementId",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"placementId": "test123"}}},
			want:    json.RawMessage(GetJSON(JSONObject{"placementId": "test123"})),
			wantErr: false,
		},
		{
			name:    "Invalid Scenerio only floorPrice is present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"floorPrice": 0.9292}}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builderApacdex(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("builderApacdex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			AssertJSON(t, tt.want, got)
		})
	}
}

func Test_builderUnruly(t *testing.T) {
	type args struct {
		params BidderParameters
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name:    "Valid Scenerio (oneOf siteId or siteid) is present-siteId",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"siteId": 123}}},
			want:    json.RawMessage(`{"siteId": 123}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf siteId or siteid) is present-siteid",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"siteid": 123}}},
			want:    json.RawMessage(`{"siteid": 123}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf siteId or siteid) is present with Optional field featureoverides",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"siteid": 123, "featureOverrides": JSONObject{"canRunUnmissable": true}}}},
			want:    json.RawMessage(`{"siteid": 123, "featureOverrides": {"canRunUnmissable": true}}`),
			wantErr: false,
		},
		{
			name:    "Invalid Scenerio (None Of siteId or siteid) is present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{}}},
			want:    json.RawMessage(``),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builderUnruly(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("builderUnruly() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			AssertJSON(t, tt.want, got)
		})
	}
}

func Test_builderBoldwin(t *testing.T) {
	type args struct {
		params BidderParameters
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name:    "Valid Scenerio (oneOf placementId or endpointId) is present-placementId",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"placementId": "1234"}}},
			want:    json.RawMessage(`{"placementId": "1234"}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf placementId or endpointId) is present-endpointId",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"endpointId": "0"}}},
			want:    json.RawMessage(`{"endpointId": "0"}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf  placementId or endpointId), Both are present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"endpointId": "0", "placementId": "1234"}}},
			want:    json.RawMessage(`{"placementId": "1234"}`),
			wantErr: false,
		},
		{
			name:    "Invalid Scenerio (None Of placementId or endpointId) is present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{}}},
			want:    json.RawMessage(``),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builderBoldwin(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("builderBoldwin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestBuilderColossus(t *testing.T) {
	type args struct {
		params BidderParameters
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name:    "Valid Scenerio (oneOf TagID or groupId) is present-TagID",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"TagID": "0"}}},
			want:    json.RawMessage(`{"TagID": "0"}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf TagID or groupId) is present-groupId",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"groupId": "0"}}},
			want:    json.RawMessage(`{"groupId": "0"}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf  TagID or groupId), Both are present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"groupId": "0", "TagID": "0"}}},
			want:    json.RawMessage(`{"TagID": "0"}`),
			wantErr: false,
		},
		{
			name:    "Invalid Scenerio (None Of TagID or groupId) is present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{}}},
			want:    json.RawMessage(``),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builderColossus(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("builderColossus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestBuilderNextmillennium(t *testing.T) {
	type args struct {
		params BidderParameters
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name: "Valid Scenerio (anyOf placement_id or group_id) is present-placement_id",
			args: args{
				params: BidderParameters{FieldMap: JSONObject{"placement_id": "1234"}},
			},
			want:    json.RawMessage(`{"placement_id":"1234"}`),
			wantErr: false,
		},
		{
			name: "Valid Scenerio (anyOf placement_id or group_id) is present-group_id",
			args: args{
				params: BidderParameters{FieldMap: JSONObject{"group_id": "1234"}},
			},
			want:    json.RawMessage(`{"group_id":"1234"}`),
			wantErr: false,
		},
		{
			name: "Valid Scenerio (anyOf placement_id or group_id) both are present",
			args: args{
				params: BidderParameters{FieldMap: JSONObject{"placement_id": "1234", "group_id": "45567"}},
			},
			want:    json.RawMessage(`{"placement_id":"1234"}`),
			wantErr: false,
		},
		{
			name:    "Invalid Scenerio (None Of placement_id or group_id) is present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{}}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builderNextmillennium(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("builderNextmillennium() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuilderRise(t *testing.T) {
	type args struct {
		params BidderParameters
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name:    "Valid Scenerio (oneOf org or publisher_id) is present-org",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"org": "0"}}},
			want:    json.RawMessage(`{"org": "0"}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf org or publisher_id) is present-publisher_id",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"publisher_id": "0"}}},
			want:    json.RawMessage(`{"publisher_id": "0"}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf org or publisher_id), Both are present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"publisher_id": "0", "org": "0"}}},
			want:    json.RawMessage(`{"org": "0"}`),
			wantErr: false,
		},
		{
			name:    "Invalid Scenerio (None Of org or publisher_id) is present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{}}},
			want:    json.RawMessage(``),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builderRise(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("builderRise() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			AssertJSON(t, tt.want, got)
		})
	}
}

func Test_builderKargo(t *testing.T) {
	type args struct {
		params BidderParameters
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name:    "Valid Scenerio (oneOf placementId or adSlotID) is present-placementId",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"placementId": "dbdsfh"}}},
			want:    json.RawMessage(`{"placementId": "dbdsfh"}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf placementId or adSlotID) is present-adSlotID",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"adSlotID": "dbdsfh"}}},
			want:    json.RawMessage(`{"adSlotID": "dbdsfh"}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf placementId or adSlotID), Both are present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"placementId": "sdhks", "adSlotID": "sdjksd"}}},
			want:    json.RawMessage(`{"placementId": "sdhks"}`),
			wantErr: false,
		},
		{
			name:    "Invalid Scenerio (None Of placementId or adSlotID) is present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{}}},
			want:    json.RawMessage(``),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builderKargo(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("builderKargo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			AssertJSON(t, tt.want, got)
		})
	}
}

func Test_builderPGAMSSP(t *testing.T) {
	type args struct {
		params BidderParameters
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name:    "Valid Scenerio (oneOf placementId or endpointId) is present-placementId",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"placementId": "dbdsfh"}}},
			want:    json.RawMessage(`{"placementId": "dbdsfh"}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf placementId or endpointId) is present-endpointId",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"endpointId": "dbdsfh"}}},
			want:    json.RawMessage(`{"endpointId": "dbdsfh"}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf placementId or endpointId), Both are present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"placementId": "sdhks", "endpointId": "sdjksd"}}},
			want:    json.RawMessage(`{"placementId": "sdhks"}`),
			wantErr: false,
		},
		{
			name:    "Invalid Scenerio (None Of placementId or endpointId) is present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{}}},
			want:    json.RawMessage(``),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builderPGAMSSP(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("builderPGAMSSP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			AssertJSON(t, tt.want, got)
		})
	}
}

func Test_builderAidem(t *testing.T) {
	type args struct {
		params BidderParameters
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name:    "Valid Scenerio rateLimit is present along with all other parameters",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"placementId": "ABCDEF", "siteId": "ABCDEF", "publisherId": "5890", "rateLimit": 0.6}}},
			want:    json.RawMessage(`{"placementId": "ABCDEF", "siteId": "ABCDEF", "publisherId": "5890"}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio rateLimit is absent along with all other parameters",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"placementId": "ABCDEF", "siteId": "ABCDEF", "publisherId": "5890"}}},
			want:    json.RawMessage(`{"placementId": "ABCDEF", "siteId": "ABCDEF", "publisherId": "5890"}`),
			wantErr: false,
		},
		{
			name:    "Invalid Scenerio (None Of publisherId or siteId) is present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{}}},
			want:    json.RawMessage(``),
			wantErr: true,
		},
		{
			name:    "Invalid Scenerio (Only publisherId) is present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"publisherId": "5890"}}},
			want:    json.RawMessage(``),
			wantErr: true,
		},
		{
			name:    "Invalid Scenerio (Only siteId) is present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"siteId": "abcd"}}},
			want:    json.RawMessage(``),
			wantErr: true,
		},
		{
			name:    "Invalid Scenerio (Only placementId) is present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"placementId": "abcd"}}},
			want:    json.RawMessage(``),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builderAidem(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("builderAidem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			AssertJSON(t, tt.want, got)
		})
	}
}

func TestBuilderCompass(t *testing.T) {
	type args struct {
		params BidderParameters
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name:    "Valid Scenerio (oneOf placementId or endpointId) is present-placementId",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"placementId": "dbdsfh"}}},
			want:    json.RawMessage(`{"placementId": "dbdsfh"}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf placementId or endpointId) is present-endpointId",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"endpointId": "dbdsfh"}}},
			want:    json.RawMessage(`{"endpointId": "dbdsfh"}`),
			wantErr: false,
		},
		{
			name:    "Valid Scenerio (oneOf placementId or endpointId), Both are present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{"placementId": "sdhks", "endpointId": "sdjksd"}}},
			want:    json.RawMessage(`{"placementId": "sdhks"}`),
			wantErr: false,
		},
		{
			name:    "Invalid Scenerio (None Of placementId or endpointId) is present",
			args:    args{params: BidderParameters{FieldMap: JSONObject{}}},
			want:    json.RawMessage(``),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builderCompass(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("builderCompass() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			AssertJSON(t, tt.want, got)
		})
	}
}
