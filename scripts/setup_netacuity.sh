#!/bin/bash

rm -rf go-netacuity-client
git clone git@git.pubmatic.com:PubMatic/go-netacuity-client.git

output=""
# get directory to get the go-netacuity-client/include directory which contains .c and .hpp files
includeDir="./go-netacuity-client/sourcecode/include"
if [ -d "$includeDir" ]; then
    output=`realpath $includeDir`
fi

echo "$output"
