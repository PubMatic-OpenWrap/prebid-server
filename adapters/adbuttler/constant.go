package adbuttler

const (
	BD_ZONE_ID             = "catalogZone"
	BD_ACCOUNT_ID          = "accountID"
	SEARCHTYPE_DEFAULT     = "exact"
	SEARCHTYPE_EXACT       = "exact"
	SEARCHTYPE_BROAD       = "broad"
	SEARCHTYPE             = "search_type"
	PAGE_SOURCE            = "page_source"
	USER_AGE               = "target_age"
	GENDER_MALE            = "Male"
	GENDER_FEMALE          = "Female"
	GENDER_OTHER           = "Others"
	DEVICE_COMPUTER        = "Personal Computer"
	DEVICE_PHONE           = "Phone"
	DEVICE_TABLET          = "Tablet"
	DEVICE_CONNECTEDDEVICE = "Connected Devices"
	USER_GENDER            = "target_gender"
	COUNTRY                = "target_country"
	REGION                 = "target_region"
	CITY                   = "target_city"
	DEVICE                 = "target_device"
	//DEFAULT_CATEGORY              = "Category"
	//DEFAULT_BRAND                 = "Brand Name"
	DEFAULT_PRODUCTID = "Product Id"
	RESPONSE_SUCCESS  = "success"
	RESPONSE_NOADS    = "NO_ADS"
	SEAT_ADBUTLER     = "adbuttler"
	BEACONTYPE_IMP    = "impression"
	BEACONTYPE_CLICK  = "click"
	IMP_KEY           = "tps_impurl="
	CLICK_KEY         = "tps_clkurl="
	CONV_HOSTNAME     = "conv_host"
	CONV_ADBUTLERID   = "conv_adbutlerID"
	CONV_ZONEID       = "conv_zoneID"
	CONV_ADBUID       = "conv_adbUID"
	CONV_IDENTIFIER   = "conv_Identifier"
	CONVERSION_URL    = `tps_ID=conv_adbutlerID&tps_setID=conv_zoneID&tps_adb_uid=conv_adbUID&tps_identifier=conv_Identifier`
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
	RESPONSE_CODE_INVALID_REQUEST = 1011
	RESPONSE_CODE_INVALID_SOURCE  = 1013
	RESPONSE_CODE_INVALID_CATALOG = 1015
	RESPONSE_CODE_UNKNOWN_ERROR   = 1210
)
