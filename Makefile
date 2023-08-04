# Makefile

all: deps test build-modules build

.PHONY: deps test build-modules build image

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
	go test github.com/prebid/prebid-server/adapters/$(adapter) -bench=.
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

mockgen: mockgeninstall mockgendb

# export GOBIN=~/go/bin; export PATH=$PATH:$GOBIN
mockgeninstall:
	go install github.com/golang/mock/mockgen@v1.6.0

mockgendb:
	mkdir -p modules/pubmatic/openwrap/database/mock modules/pubmatic/openwrap/database/mock_driver
	mockgen database/sql/driver Driver,Connector,Conn,DriverContext > modules/pubmatic/openwrap/database/mock_driver/mock.go
	mockgen github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/database Database > modules/pubmatic/openwrap/database/mock/mock.go

mockgencache:
	mkdir -p modules/pubmatic/openwrap/cache/mock
	mockgen github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/cache Cache > modules/pubmatic/openwrap/cache/mock/mock.go