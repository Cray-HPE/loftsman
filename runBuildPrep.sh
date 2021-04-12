#!/bin/bash

set -e

GO_VERSION="1.16.2"
INSTALLED_GO_VERSION=$(go version | awk '{print $3}')
HELM_VERSION="v3.2.4"

if [[ "go${GO_VERSION}" !=  $INSTALLED_GO_VERSION ]]; then
    echo "Upgrading go from version ${INSTALLED_GO_VERSION} to ${GO_VERSION}"
    go get golang.org/dl/go$GO_VERSION || true
    $GOPATH/bin/go$GO_VERSION download || true
    GO_EXEC=$(which go)
    rm -f $GO_EXEC
    cp $GOPATH/bin/go$GO_VERSION $GO_EXEC
fi

mkdir -p $GOPATH/bin
mkdir -p $GOPATH/src
mkdir -p $GOPATH/pkg

if ! command -v bc &> /dev/null; then
    zypper -n install -y bc
fi

echo "Getting the Helm binary to package with the rpm"
# we need this helm v3 binary to be on the path for tests
wget -q https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm
# and we'll put it in the root here as well to package with the rpm
cp /usr/local/bin/helm ./helm

echo "Running tests"
./scripts/tests-unit.sh

echo "Building loftsman binaries"
mkdir -p .build
git_version=$(git rev-parse --short HEAD)
go build -o ./.build/loftsman-linux-amd64 -ldflags "-X 'github.com/Cray-HPE/loftsman/cmd.Version=${git_version}'"

VERSION=$(./.build/loftsman-linux-amd64 --version | awk '{print $3}')
if [ ! -z "${BUILD_NUMBER}" ]; then
    VERSION="${VERSION}.${BUILD_NUMBER}"
fi
# Add build number to version
echo "$VERSION" > ./.version
