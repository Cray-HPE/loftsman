#!/bin/bash

RPM=$(ls -l RPMS |grep rpm | grep -v src | awk '{print $NF}')

if command -v yum > /dev/null; then
    yum install -y RPMS/$RPM
elif command -v zypper > /dev/null; then
    zypper --no-gpg-checks install -y -f -l RPMS/$RPM
else
    echo "Unsupported package manager or package manager not found -- installing nothing"
    exit 1
fi

set -e

. smokeTests.sh