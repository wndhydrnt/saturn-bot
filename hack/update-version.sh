#!/usr/bin/env bash

current_version="$1"
next_version="$2"

sed_exec="sed"
if [[ "$(uname -s)" == "Darwin" ]]; then
  sed_exec="gsed"
fi

if ! command -v ${sed_exec} &> /dev/null; then
    echo "sed isn't available. On macos install GNU sed: brew install gnu-sed"
    exit 1
fi

echo "Replacing versions with ${sed_exec}"

${sed_exec} -i -e "s/version: ${current_version}/version: ${next_version}/g" "pkg/server/api/openapi/openapi.yaml"

${sed_exec} -i -e "s/${current_version}/${next_version}/g" "docs/installation.md"

${sed_exec} -i -e "s/${current_version}/${next_version}/g" "deploy/docker-compose/docker-compose.yml"
