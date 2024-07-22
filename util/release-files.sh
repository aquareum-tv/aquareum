#!/bin/bash

# gets all the files for a release version

set -euo pipefail
set -x

VERSION="$1"

id="$(curl --silent --fail "$CI_API_V4_URL/projects/$CI_PROJECT_ID/packages?package_version=$VERSION" | jq -r '.[0].id')"
files="$(curl --silent --fail "$CI_API_V4_URL/projects/$CI_PROJECT_ID/packages/$id/package_files")"
echo $files | jq -c '[.[] | {"name": .file_name, "url": ("'$CI_API_V4_URL'/projects/'$CI_PROJECT_ID'/packages/generic/latest/'$VERSION'/" + .file_name)}]'