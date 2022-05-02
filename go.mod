module github.com/PubMatic-OpenWrap/prebid-server

go 1.16

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/NYTimes/gziphandler v1.1.1
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/beevik/etree v1.0.2
	github.com/buger/jsonparser v1.1.1
	github.com/chasex/glog v0.0.0-20160217080310-c62392af379c
	github.com/coocood/freecache v1.2.0
	github.com/docker/go-units v0.4.0
	github.com/gofrs/uuid v4.2.0+incompatible
	github.com/golang/glog v1.0.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/lib/pq v1.0.0
	github.com/magiconair/properties v1.8.5
	github.com/mitchellh/copystructure v1.2.0
	github.com/mxmCherry/openrtb/v15 v15.0.1
	github.com/prebid/go-gdpr v1.11.0
	github.com/prometheus/client_golang v1.12.1
	github.com/prometheus/client_model v0.2.0
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/rs/cors v1.8.2
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/go-metrics-influxdb v0.1.1
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20180816142147-da425ebb7609
	github.com/yudai/gojsondiff v1.0.0
	github.com/yudai/golcs v0.0.0-20170316035057-ecda9a501e82 // indirect
	github.com/yudai/pp v2.0.1+incompatible // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/text v0.3.7
	gopkg.in/evanphx/json-patch.v4 v4.12.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/prebid/prebid-server => ./

replace github.com/mxmCherry/openrtb/v15 => github.com/PubMatic-OpenWrap/openrtb/v15 v15.0.0-20210514055459-92ccbf3eb6fe

replace github.com/beevik/etree v1.0.2 => github.com/PubMatic-OpenWrap/etree v1.0.2-0.20210129100623-8f30cfecf9f4
