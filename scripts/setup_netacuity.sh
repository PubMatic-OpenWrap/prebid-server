#!/bin/bash

# download mod dependencies
go mod tidy

# list of directories to search for go-netacuity-client pkg
directories=("$GOPATH" "vendor" "../../../go/pkg/mod")

netacuityDir=""
for dir in "${directories[@]}"; do
    if [ -d "$dir" ]; then
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

echo $includeDir
