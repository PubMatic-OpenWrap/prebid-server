#!/bin/bash

set -e

RACE=0
AUTOFMT=true
COVERAGE=false
VET=true

while true; do
  case "$1" in
     --nofmt ) AUTOFMT=false; shift ;;
     --race ) RACE=$2; shift; shift; ;;
     --cov ) COVERAGE=true; shift ;;
     --novet ) VET=false; shift ;;
     * ) break ;;
  esac
done


set_cflag_for_netacuity() {
    go mod tidy
    netacuityDir=`find ../../../go/pkg/mod -type d -iname 'go-netacuity-client@*'`
    echo "netacuityDir=$netacuityDir"
    includeDir=`find $netacuityDir -type d -iname include | xargs realpath`
    echo "includeDir=$includeDir"
    export CGO_CFLAGS="-I $includeDir"
    echo "CGO_CFLAGS=$CGO_CFLAGS"
}


set_cflag_for_netacuity
./scripts/format.sh -f $AUTOFMT


# Run the actual tests. Make sure there's enough coverage too, if the flags call for it.
if $COVERAGE; then
  ./scripts/check_coverage.sh
else
  go test -timeout 120s $(go list ./... | grep -v /vendor/)
fi

# Then run the race condition tests. These only run on tests named TestRace.* for two reasons.
#
#   1. To speed things up (for large -count values)
#   2. Because some tests open up files on the filesystem, and some operating systems limit the number of open files for a single process.
if [ "$RACE" -ne "0" ]; then
  echo "time to run go race"
  go test -race $(go list ./... | grep -v /vendor/) -run ^TestRace.*$ -count $RACE
fi

if $VET; then
  echo "Running go vet check"
  # go vet -composites=false ./...
fi
