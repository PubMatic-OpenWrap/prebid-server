# Makefile

all: deps test build-modules build

.PHONY: deps test build-modules build image format

# deps will clean out the vendor directory and use go mod for a fresh install
deps:
	GOPROXY="https://proxy.golang.org" go mod vendor -v && go mod tidy -v

# test will ensure that all of our dependencies are available and run validate.sh
test: deps
# If there is no indentation, Make will treat it as a directive for itself; otherwise, it's regarded as a shell script.
# https://stackoverflow.com/a/4483467
ifeq "$(adapter)" ""
	./validate.sh
else
	go test github.com/prebid/prebid-server/v3/adapters/$(adapter) -bench=.
endif

# build-modules generates modules/builder.go file which provides a list of all available modules
build-modules:
	go generate modules/modules.go

# build will ensure all of our tests pass and then build the go binary
build: test
	go build -mod=vendor ./...

# image will build a docker image
image:
	docker build -t prebid-server .

# format runs format
format:
	./scripts/format.sh -f true

# formatcheck runs format for diagnostics, without modifying the code
formatcheck:
	./scripts/format.sh -f false

mockgen: mockgeninstall mockgendb mockgencache mockgenmetrics mockgenlogger mockgenpublisherfeature mockgenprofilemetadata mockgenwakanda

# export GOPATH=~/go ; GOBIN=~/go/bin; export PATH=$PATH:$GOBIN
mockgeninstall:
	go install github.com/golang/mock/mockgen@v1.6.0

mockgendb:
	mkdir -p modules/pubmatic/openwrap/database/mock modules/pubmatic/openwrap/database/mock_driver
	mockgen database/sql/driver Driver,Connector,Conn,DriverContext > modules/pubmatic/openwrap/database/mock_driver/mock.go
	mockgen github.com/PubMatic-OpenWrap/prebid-server/v3/modules/pubmatic/openwrap/database Database > modules/pubmatic/openwrap/database/mock/mock.go

mockgencache:
	mkdir -p modules/pubmatic/openwrap/cache/mock
	mockgen github.com/PubMatic-OpenWrap/prebid-server/v3/modules/pubmatic/openwrap/cache Cache > modules/pubmatic/openwrap/cache/mock/mock.go

mockgenmetrics:
	mkdir -p modules/pubmatic/openwrap/metrics/mock
	mockgen github.com/PubMatic-OpenWrap/prebid-server/v3/modules/pubmatic/openwrap/metrics MetricsEngine > modules/pubmatic/openwrap/metrics/mock/mock.go

mockgengeodb:
	mkdir -p modules/pubmatic/openwrap/geodb/mock
	mockgen github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/geodb Geography > modules/pubmatic/openwrap/geodb/mock/mock.go

mockgenlogger:
	mkdir -p analytics/pubmatic/mhttp/mock
	mockgen github.com/PubMatic-OpenWrap/prebid-server/v3/analytics/pubmatic/mhttp HttpCallInterface,MultiHttpContextInterface > analytics/pubmatic/mhttp/mock/mock.go

mockgenpublisherfeature:
	mkdir -p modules/pubmatic/openwrap/publisherfeature
	mockgen github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/publisherfeature Feature > modules/pubmatic/openwrap/publisherfeature/mock/mock.go

mockgenwakanda: 
	mkdir -p modules/pubmatic/openwrap/wakanda/mock
	mockgen github.com/PubMatic-OpenWrap/prebid-server/v3/modules/pubmatic/openwrap/wakanda Commands,DebugInterface > modules/pubmatic/openwrap/wakanda/mock/mock.go

mockgenprofilemetadata:
	mkdir -p modules/pubmatic/openwrap/profilemetadata/mock
	mockgen github.com/PubMatic-OpenWrap/prebid-server/v3/modules/pubmatic/openwrap/profilemetadata ProfileMetaData > modules/pubmatic/openwrap/profilemetadata/mock/mock.go
