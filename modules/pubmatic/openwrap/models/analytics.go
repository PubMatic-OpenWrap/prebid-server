package models

// Profile platform types
var ProfileTypePlatform = map[string]int{
	"display": 1,
	"amp":     2,
	"video":   3,
	"in-app":  4,
}

// App integration path
var AppIntegrationPath map[string]int = map[string]int{
	"iOS":                 1,
	"Android":             2,
	"React Native Plugin": 3,
	"Flutter Plugin":      4,
	"Unity Plugin":        5,
	"Other":               0,
}

// App sub integration path
var AppSubIntegrationPath map[string]int = map[string]int{
	"DFP":                         1,
	"MoPub":                       3,
	"CUSTOM":                      4,
	"Primary Ad Sdk":              5,
	"Google Ad Manager":           6,
	"AppLovin Max Custom Adapter": 7,
	"AppLovin Max SDK Bidding":    8,
	"IronSource":                  9,
	"AdMob":                       10,
	"Other":                       0,
}
