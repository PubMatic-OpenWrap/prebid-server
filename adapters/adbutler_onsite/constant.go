package adbutler_onsite

const (
	RESPONSE_SUCCESS   = "SUCCESS"
	RESPONSE_NOADS     = "NO_ADS"
	IMAGE_URL_TEMPLATE = "<div id='%s' style='margin:0;padding:0;'><img src='%s' width='%s' height='%s'></div>"

	Seat_AdbutlerOnsite        = "adbutler_onsite"
	InventoryIDOnsite_Prefix   = "InventoryID_"
	AdButler_Req_Type          = "json"
	AdButler_Req_Ads           = "all"
	DBAdtype_Banner            = "Static Ad"
	DBAdtype_Custom_Html       = "Dynamic Ad"
	AdButlerAdtype_Banner      = "image"
	AdButlerAdtype_Custom_Html = "raw"

	Adtype_Invalid       = 0
	Adtype_Banner        = 1
	Adtype_Custom_Banner = 2

	Ampersand = "&"

	IMP_KEY        = "tps_impurl="
	CLICK_KEY      = "tps_clkurl="
	VIEW_KEY       = "tps_vwurl="
	CreativeId_KEY = "tp_crid="

	Pattern_Click_URL = `href="(https?://[^\s]+/redirect\.spark\?[^"]+)"`
)
