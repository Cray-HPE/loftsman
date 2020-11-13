#!/bin/bash

this_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $this_dir/../

if ! command -v mockery &> /dev/null; then
  echo "ERROR: couldn't find mockery in \$PATH, cannot generate mocks"
  exit 1
fi
find ./mocks -mindepth 1 -maxdepth 1 -not -name custom-mocks -exec rm -rf '{}' \;
mv ./mocks/custom-mocks ./mocks/_tmp
mockery -all -keeptree
mv ./mocks/_tmp ./mocks/custom-mocks

# write go files to empty package directories to prevent errors/warnings
cat <<EOF > ./mocks/internal/Placeholder.go
package mocks
EOF

cd ./mocks
ln -sf ./internal/interfaces ./interfaces
