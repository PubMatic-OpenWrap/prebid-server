#!/bin/bash
# Generate test coverage statistics for Go packages.
# 
# Works around the fact that `go test -coverprofile` currently does not work
# with multiple packages, see https://code.google.com/p/go/issues/detail?id=6909
#
# Usage: script/coverage.sh [--html]
#
#     --html      Additionally create HTML report and open it in browser
#

set -e

workdir=.cover
profile="$workdir/cover.out"
mode=count

setup_netacuity() {
    echo "create netacuity dir"
    mkdir -p /usr/local/net_acuity_client/lib/
    mkdir -p /usr/local/net_acuity_client/include/

    echo "find in /go/pkg/mod"
    find /go/pkg/mod -type d -iname 'go-netacuity-client@*'

    echo "find in ~/go/pkg/mod"
    find ~/go/pkg/mod -type d -iname 'go-netacuity-client@*'
    echo "calling make on go-netacuity-client library"
    make -C `find ~/go/pkg/mod -type d -iname 'go-netacuity-client@*'`

    echo "print /usr/local/net_acuity_client"
    ls -ll /usr/local/net_acuity_client/*
    echo "removing setup_netacuity"
}

generate_cover_data() {
    rm -rf "$workdir"
    mkdir "$workdir"

    for pkg in "$@"; do
        f="$workdir/$(echo $pkg | tr / -).cover"
        cover=""
        if ! [[ "$pkg" =~ ^github\.com\/PubMatic\-OpenWrap\/prebid\-server$ ]]; then
            cover="-covermode=$mode -coverprofile=$f"
        fi
        # util/task uses _test package name
        if [[ "$pkg" =~ ^github\.com\/PubMatic\-OpenWrap\/prebid\-server\/util\/task$ ]]; then
            cover+=" -coverpkg=github.com/prebid/prebid-server/util/task"
        fi

        if [[ "$pkg" =~ ^github\.com\/PubMatic\-OpenWrap\/prebid\-server\/router$ ]]; then
            cover+=" -coverpkg=github.com/prebid/prebid-server/router"
        fi

        # temporarily disable openwrap, remove as we add full support to each package
        if [[ "$pkg" =~ ^github\.com\/PubMatic\-OpenWrap\/prebid\-server\/modules\/pubmatic\/openwrap$ ]]; then
            cover+=" -coverpkg=github.com/prebid/prebid-server/modules/pubmatic/openwrap"
        fi

        if [[ "$pkg" =~ ^github\.com\/PubMatic\-OpenWrap\/prebid\-server\/modules\/pubmatic\/openwrap\/adapters$ ]]; then
            cover+=" -coverpkg=github.com/prebid/prebid-server/modules/pubmatic/openwrap/adapters"
        fi

        if [[ "$pkg" =~ ^github\.com\/PubMatic\-OpenWrap\/prebid\-server\/modules\/pubmatic\/openwrap\/adunitconfig$ ]]; then
            cover+=" -coverpkg=github.com/prebid/prebid-server/modules/pubmatic/openwrap/adunitconfig"
        fi

        if [[ "$pkg" =~ ^github\.com\/PubMatic\-OpenWrap\/prebid\-server\/modules\/pubmatic\/openwrap\/bidderparams$ ]]; then
            cover+=" -coverpkg=github.com/prebid/prebid-server/modules/pubmatic/openwrap/bidderparams"
        fi

        if [[ "$pkg" =~ ^github\.com\/PubMatic\-OpenWrap\/prebid\-server\/modules\/pubmatic\/openwrap\/config$ ]]; then
            cover+=" -coverpkg=github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
        fi

        if [[ "$pkg" =~ ^github\.com\/PubMatic\-OpenWrap\/prebid\-server\/modules\/pubmatic\/openwrap\/database$ ]]; then
            cover+=" -coverpkg=github.com/prebid/prebid-server/modules/pubmatic/openwrap/database"
        fi

        if [[ "$pkg" =~ ^github\.com\/PubMatic\-OpenWrap\/prebid\-server\/modules\/pubmatic\/openwrap\/metrics$ ]]; then
            cover+=" -coverpkg=github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
        fi

        if [[ "$pkg" =~ ^github\.com\/PubMatic\-OpenWrap\/prebid\-server\/modules\/pubmatic\/openwrap\/metrics\/stats$ ]]; then
            cover+=" -coverpkg=github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/stats"
        fi

        if [[ "$pkg" =~ ^github\.com\/PubMatic\-OpenWrap\/prebid\-server\/modules\/pubmatic\/openwrap\/models$ ]]; then
            cover+=" -coverpkg=github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
        fi

        if [[ "$pkg" =~ ^github\.com\/PubMatic\-OpenWrap\/prebid\-server\/modules\/pubmatic\/openwrap\/geodb$ ]]; then
            cover+=" -coverpkg=github.com/prebid/prebid-server/modules/pubmatic/openwrap/geodb"
        fi

        go mod download all
        ls -ll ../../../go/pkg/mod/git.pubmatic.com/!pub!matic/*/*
        netacuityDir=`find ../../../go/pkg/mod -type d -iname 'go-netacuity-client@*'`
        echo "netacuityDir=$netacuityDir"
        includeDir=`find $netacuityDir -type d -iname include`
        export CGO_CFLAGS="-I $includeDir"
        echo "CGO_CFLAGS=$CGO_CFLAGS"
        
        go test  ${cover} "$pkg"
        #go test -tag exclude_feature ${cover} "$pkg"
    done

    echo "mode: $mode" >"$profile"
    grep -h -v "^mode:" "$workdir"/*.cover >>"$profile"
}

show_cover_report() {
    go tool cover -${1}="$profile"
}

generate_cover_data $(go list ./... | grep -v /vendor/)
#generate_cover_data $(go list ./... | grep modules)
#show_cover_report func
case "$1" in
"")
    ;;
--html)
    show_cover_report html ;;
*)
    echo >&2 "error: invalid option: $1"; exit 1 ;;
esac
