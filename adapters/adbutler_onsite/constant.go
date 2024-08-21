package adbutler_onsite

const (
	RESPONSE_SUCCESS   = "SUCCESS"
	RESPONSE_NOADS     = "NO_ADS"
	IMAGE_URL_TEMPLATE = "<div id='%s' style='margin:0;padding:0;'><img src='%s' width='%s' height='%s'></div>"

	Seat_AdbutlerOnsite        = "adbutler_onsite"
	InventoryIDOnsite_Prefix   = "InventoryID_"
	AdButler_Req_Type          = "json"
	AdButler_Req_Ads           = "all"
	DBAdtype_Banner            = "Banner"
	DBAdtype_Custom_Html       = "Native"
	AdButlerAdtype_Banner      = "image"
	AdButlerAdtype_Custom_Html = "raw"

	Adtype_Invalid       = 0
	Adtype_Banner        = 1
	Adtype_Custom_Banner = 2

	IMP_KEY   = "tps_impurl="
	CLICK_KEY = "tps_clkurl="
	VIEW_KEY  = "tps_vwurl="
)
