package models

import "github.com/prebid/prebid-server/v2/openrtb_ext"

type ExtSource struct {
	*openrtb_ext.ExtSource
	OMIDPV string `json:"omidpv,omitempty"`
	OMIDPN string `json:"omidpn,omitempty"`
}
