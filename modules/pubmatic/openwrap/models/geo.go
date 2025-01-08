package models

import "time"

// compliance consts
const (
	GDPRCompliance      = "GDPR"
	USPCompliance       = "USP"
	GPPCompliance       = "GPP"
	CountryCodeUS       = "US"
	StateCodeCalifornia = "ca"
)

// headers const
const (
	CacheTimeout                        = time.Duration(48) * time.Hour
	HeaderContentType                   = "Content-Type"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderCacheControl                  = "Cache-Control"
	HeaderContentTypeValue              = "application/json"
	HeaderAccessControlAllowOriginValue = "*"
	HeaderOriginKey                     = "Origin"
	HeaderRefererKey                    = "Referer"
)
