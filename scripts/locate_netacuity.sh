#!/bin/bash

set -e
whoami
# download mod dependencies
go mod tidy

echo "GOPATH=$GOPATH"

# list of directories to search for go-netacuity-client pkg
directories=("$GOPATH/pkg/mod" "~/go/pkg/mod" "vendor" "../../../go/pkg/mod")

netacuityDir=""
for dir in "${directories[@]}"; do
    echo "dir=$dir"
    if [ -d "$dir" ]; then
        echo "dir=$dir exist"
        netacuityDir=`find "$dir" -type d -iname 'go-netacuity-client@*'`
        if [ "$netacuityDir" != "" ];then
            break
        fi
    fi
done

# get directory to get the go-netacuity-client/include directory which contains .c and .hpp files
includeDir=""
if [ -d "$netacuityDir" ]; then
    includeDir=`find $netacuityDir -type d -iname include | xargs realpath`
fi

# if includeDir does not exist then exit with failure state
if [ ! -d "$includeDir" ]; then
    exit 1
fi

echo "includeDir=$includeDir"
