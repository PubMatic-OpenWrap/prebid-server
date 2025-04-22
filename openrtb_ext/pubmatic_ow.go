package openrtb_ext

type BuyerCreative struct {
	BuyerCreativeId string `json:"buyer_creative_id,omitempty"`
}

type GoogleSDKParams struct {
	BillingIds                []int64         `json:"billing_id,omitempty"`
	PublisherSettingListIds   []int64         `json:"publisher_setting_list_id,omitempty"`
	AllowedVendorType         []int32         `json:"allowed_vendor_type,omitempty"`
	ExcludedCreatives         []BuyerCreative `json:"excluded_creatives,omitempty"`
	IsAppOpenAd               bool            `json:"is_app_open_ad,omitempty"`
	AllowedRestrictedCategory int32           `json:"allowed_restricted_category,omitempty"`
}
