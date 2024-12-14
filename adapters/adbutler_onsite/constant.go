package adbutler_onsite

const (
	RESPONSE_SUCCESS   = "SUCCESS"
	RESPONSE_NOADS     = "NO_ADS"
	IMAGE_URL_TEMPLATE = "<div style='margin:0;padding:0;'><a href='ACTUAL_CLICK_URL'><img src='%s'></a></div>"
	IMAGE_URL_TEMPLATE_TARGET = "<div style='margin:0;padding:0;'><a href='ACTUAL_CLICK_URL' target='REDIRECT_TARGET'><img src='%s'></a></div>"
	DYNAMIC_IMAGE_URL_CLICKTEMPLATE = "<script>document.getElementById('IMPRESSION_ID').addEventListener('click', function handleAdClick(event){if(!event.target.closest('a')){window.open('ACTUAL_CLICK_URL');}}</script>"
	DYNAMIC_IMAGE_URL_AHREFSTART = "<a href='ACTUAL_CLICK_URL'>"
	DYNAMIC_IMAGE_URL_AHREFSTART_TARGET = "<a href='ACTUAL_CLICK_URL' target='REDIRECT_TARGET'>"
	DYNAMIC_IMAGE_URL_AHREFEND   = "</a>"
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

	IMP_KEY   = "tps_impurl="
	CLICK_KEY = "tps_clkurl="
	VIEW_KEY  = "tps_vwurl="

	Pattern_Click_URL = `href="(https?://[^\s]+/redirect\.spark\?[^"]+)"`
)



