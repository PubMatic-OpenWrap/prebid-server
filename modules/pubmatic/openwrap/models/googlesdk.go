package models

import (
	nativeResponse "github.com/prebid/openrtb/v20/native1/response"
)

type GoogleSDKBidExt struct {
	SDKRenderedAd          SDKRenderedAd           `json:"sdk_rendered_ad,omitempty"`
	EventNotificationToken *EventNotificationToken `json:"event_notification_token,omitempty"`
	BillingID              string                  `json:"billing_id,omitempty"`
	ProcessingTime         int                     `json:"processing_time_ms,omitempty"`
}

type SDKRenderedAd struct {
	ID            string     `json:"id,omitempty"`
	RenderingData string     `json:"rendering_data,omitempty"`
	DeclaredAd    DeclaredAd `json:"declared_ad,omitempty"`
}

type DeclaredAd struct {
	ClickThroughURL []string                 `json:"click_through_url,omitempty"`
	HTMLSnippet     string                   `json:"html_snippet,omitempty"`
	VideoVastXML    string                   `json:"video_vast_xml,omitempty"`
	NativeResponse  *nativeResponse.Response `json:"native_response,omitempty"`
}

type EventNotificationToken struct {
	Payload string `json:"payload,omitempty"`
}
