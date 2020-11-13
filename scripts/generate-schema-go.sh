#!/bin/bash

this_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $this_dir/../

for schema_json_file in $(find . -name schema.json); do
  dir=$(dirname $schema_json_file)
  if [ -f $dir/schema.go ]; then
    package_name=$(cat $dir/schema.go | grep -e ^package)
    cat > ${dir}/schema.json.go <<EOF
$package_name

// AUTO-GENERATED FILE: DO NOT MODIFY

// Schema is the Go string variable container the JSON schema
const Schema = \`
$(cat $schema_json_file)
\`
EOF
  fi
done
