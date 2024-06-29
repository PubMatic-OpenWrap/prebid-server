package adbutler_onsite

const (
	RESPONSE_SUCCESS   = "SUCCESS"
	RESPONSE_NOADS     = "NO_ADS"
	IMAGE_URL_TEMPLATE = "<div id='%s' style='margin:0;padding:0;'><img src='%s' width='%s' height='%s'></div>"

	INVALID_ADTYPE = 0
	BANNER_ADTYPE  = 1
	NATIVE_ADTYPE  = 2

	Seat_AdbutlerOnsite        = "adbutler_onsite"
	InventoryIDOnsite_Prefix   = "InventoryID_"
	AdButler_Req_Type          = "json"
	AdButler_Req_Ads           = "all"
	DBAdtype_Banner            = "Banner"
	DBAdtype_Custom_Html       = "Native"
	AdButlerAdtype_Banner      = "image"
	AdButlerAdtype_Custom_Html = "raw"

	RequestAdtype_Banner      = 1
	RequestAdtype_Custom_Html = 2
)
