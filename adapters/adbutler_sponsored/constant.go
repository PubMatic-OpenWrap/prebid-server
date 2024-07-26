package adbutler_sponsored

const (
	BD_ZONE_ID         = "catalogZone"
	BD_ACCOUNT_ID      = "accountID"
	SEARCHTYPE_DEFAULT = "exact"
	SEARCHTYPE_EXACT   = "exact"
	SEARCHTYPE_BROAD   = "broad"
	SEARCHTYPE         = "search_type"
	PAGE_SOURCE        = "page_source"

	//DEFAULT_CATEGORY              = "Category"
	//DEFAULT_BRAND                 = "Brand Name"
	DEFAULT_PRODUCTID  = "Product Id"
	RESPONSE_SUCCESS   = "success"
	RESPONSE_NOADS     = "NO_ADS"
	SEAT_ADBUTLER      = "adbutler_sponsored"
	BEACONTYPE_IMP     = "impression"
	BEACONTYPE_ELG_IMP = "eligible_impression"
	BEACONTYPE_CLICK   = "click"
	IMP_KEY            = "tps_impurl="
	CLICK_KEY          = "tps_clkurl="
	CONV_HOSTNAME      = "conv_host"
	CONV_ADBUTLERID    = "conv_adbutlerID"
	CONV_ZONEID        = "conv_zoneID"
	CONV_ADBUID        = "conv_adbUID"
	CONV_IDENTIFIER    = "conv_Identifier"
	CONVERSION_URL     = `tps_ID=conv_adbutlerID&tps_setID=conv_zoneID&tps_adb_uid=conv_adbUID&tps_identifier=conv_Identifier`
	//PD_TEMPLATE_BRAND             = "brandName"
	//PD_TEMPLATE_CATEGORY          = "categories"
	PD_TEMPLATE_PRODUCTID     = "productId"
	PD_TEMPLATE_SUBCATEGORY   = "subcategories"
	ProductTemplate_Separator = "$##$"
	DATATYE_NUMBER            = 1
	DATATYE_STRING            = 2
	DATATYE_ARRAY             = 3
	DATATYE_DATE              = 4
	DATATYE_TIME              = 5
	DATATYE_DATETIME          = 6
)

const (
	ADBUTLER_RESPONSE_CODE_INVALID_REQUEST = 1011
	ADBUTLER_RESPONSE_CODE_INVALID_SOURCE  = 1013
	ADBUTLER_RESPONSE_CODE_INVALID_CATALOG = 1015
	ADBUTLER_RESPONSE_CODE_UNKNOWN_ERROR   = 1210
)
