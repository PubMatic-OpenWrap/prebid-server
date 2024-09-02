package wakanda

var (

	//contentType HTTP Response Header Content Type
	contentType = `Content-Type`
	//contentTypeApplicationJSON HTTP Header Content-Type Value
	contentTypeApplicationJSON = `application/json`
)

const (
	//cMaxTraceCount maximum trace request can be logged
	cMaxTraceCount = 20
	//cAPIDebugLevel debug level parameter of wakanda handler
	cAPIDebugLevel = "debugLevel"
	//cAPIPublisherID publisher id paramater of wakanda handler
	cAPIPublisherID = "pubId"
	//cAPIProfileID profile id parameter of wakanda handler
	cAPIProfileID = "profId"
	//cRuleKeyPubProfile rule format ,same is used for folder name with "__DC"
	cRuleKeyPubProfile = "PUB:%s__PROF:%s"
	cAPIFilters        = "filters"
)
