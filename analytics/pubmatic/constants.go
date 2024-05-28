package pubmatic

// Profile types
const (
	openWrap  = 1
	identity  = 2
	prebidS2S = 3
)

// Profile platform types
var profileTypePlatform = map[string]int{
	"display": 1,
	"amp":     2,
	"video":   3,
	"in-app":  4,
}

var appPlatform = map[string]int{
	"iOS":     4,
	"Android": 5,
}

// App integration path
var appIntegrationPath map[string]int = map[string]int{
	"Android":             2,
	"React Native Plugin": 3,
	"Flutter Plugin":      4,
	"Unity Plugin":        5,
	"Other":               0,
}

// App sub integration path
var appSubIntegrationPath map[string]int = map[string]int{
	"DFP":                         1,
	"MoPub":                       3,
	"CUSTOM":                      4,
	"Primary Ad Sdk":              5,
	"Google Ad Manager":           6,
	"AppLovin Max Custom Adapter": 7,
	"AppLovin Max SDK Bidding	":   8,
	"IronSource":                  9,
	"AdMob":                       10,
	"Other":                       0,
}
