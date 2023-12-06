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
e
# Setup netacuity directory and use the location to set the CGO_CFLAG
NETACUITY_DIR=`./scripts/setup_netacuity.sh`
if [[ "$?" -ne "0" ]]; then
  echo "NETACUITY_DIR=$NETACUITY_DIR"
  echo -e "Fail to set the CGO_CFLAG for go-netacuity-client repository"
  exit 1
fi
echo "NETACUITY_DIR=$NETACUITY_DIR"
export CGO_CFLAGS="-I $NETACUITY_DIR"

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
  go test -race $(go list ./... | grep -v /vendor/) -run ^TestRace.*$ -count $RACE
fi

if $VET; then
  echo "Running go vet check"
  go vet -composites=false ./...
fi
