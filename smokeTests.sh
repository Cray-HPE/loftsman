#!/bin/bash

# Run some sanity tests to make sure the built binary works.
CLI="./.build/loftsman-linux-amd64"
VERSION=$(cat ./.version)

cli_help=$($CLI --help)
if [[ $? == 0 ]]; then
    echo "PASS: loftsman returns help"
else
    echo "FAIL: loftsman returns an error."
    exit 1
fi


cli_version=$($CLI --version | awk '{print $3}')
if [ ! -z "${BUILD_NUMBER}" ]; then
    cli_version="${cli_version}.${BUILD_NUMBER}"
fi
if [[ $? == 0 ]]; then
    echo "PASS: loftsman version returned"
else
    echo "FAIL: loftsman version fails"
    exit 1
fi

if [[ $VERSION == $cli_version ]]; then
    echo "PASS: loftsman version successfully generated"
else
    echo "FAIL: loftsman version does not match build version. Expected: '${VERSION}', got: '${cli_version}'"
    exit 1
fi
