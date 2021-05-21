#!/bin/bash

set -e
set -o pipefail

minimum_coverage_percentage=80

this_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
pushd /tmp

if ! command -v golint &>/dev/null; then
  go get -u golang.org/x/lint/golint
fi
golint="golint"
if ! command -v golint &>/dev/null; then
  golint="$GOPATH/bin/golint"
fi

pushd $this_dir/../
lint_result=$($golint ./...)
if [[ ! -z "$lint_result" ]]; then
  echo "Lint failed: $lint_result"
  exit 1
fi

if ! command -v bc &> /dev/null; then
  echo "ERROR: couldn't find bc in \$PATH, required for running tests"
  exit 1
fi

mkdir -p ./.tests
if ! go test -count=1 -v -coverprofile=./.tests/coverage.out ./... | tee ./.tests/tests-unit.log; then
  echo ""
  echo "ERROR: Unit tests failed"
  echo ""
  exit 1
fi
go tool cover -html=./.tests/coverage.out -o ./.tests/coverage.html
total_percent=0
total_count=0
coverage_results=$(cat ./.tests/tests-unit.log | grep coverage: | awk -F'coverage: ' '{print $2}' | awk -F'%' '{print $1}')
for perc in $coverage_results; do
  total_percent=$(echo "$total_percent + $perc" | bc)
  let total_count=$(($total_count + 1))
done
aggregated_percentage=$(echo "$total_percent/$total_count" | bc)
echo "Total percent coverage: ${aggregated_percentage}%" | tee -a ./.tests/tests-unit.log
if (( $(echo "$aggregated_percentage < $minimum_coverage_percentage" | bc -l) )); then
  echo "ERROR: minimum code coverage of ${minimum_coverage_percentage}% not met: at ${aggregated_percentage}% coverage"
  exit 1
fi
