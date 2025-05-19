package openrtb_ext

type BuyerCreative struct {
	BuyerCreativeId string `json:"buyer_creative_id,omitempty"`
}

type CreativeEnforcementSettings struct {
	PolicyEnforcement          int `json:"policy_enforcement,omitempty"`
	ScanEnforcement            int `json:"scan_enforcement,omitempty"`
	PublisherBlocksEnforcement int `json:"publisher_blocks_enforcement,omitempty"`
}

type GoogleSDKParams struct {
	BillingIds                  []string                     `json:"billing_id,omitempty"`
	PublisherSettingListIds     []string                     `json:"publisher_setting_list_id,omitempty"`
	AllowedVendorType           []int                        `json:"allowed_vendor_type,omitempty"`
	ExcludedCreatives           []BuyerCreative              `json:"excluded_creatives,omitempty"`
	IsAppOpenAd                 int8                         `json:"is_app_open_ad,omitempty"`
	AllowedRestrictedCategory   []int                        `json:"allowed_restricted_category,omitempty"`
	CreativeEnforcementSettings *CreativeEnforcementSettings `json:"creative_enforcement_settings,omitempty"`
}

type ExtImpBanner struct {
	Flexslot *struct {
		Wmin int32 `json:"wmin,omitempty"`
		Wmax int32 `json:"wmax,omitempty"`
		Hmin int32 `json:"hmin,omitempty"`
		Hmax int32 `json:"hmax,omitempty"`
	} `json:"flexslot,omitempty"`
}
