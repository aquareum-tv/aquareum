#!/bin/bash

# prints out current version from git tags

set -euo pipefail

if [[ ${CI_COMMIT_TAG:-} != "" ]]; then
  echo "$CI_COMMIT_TAG"
else
  git describe --long --tags --dirty | sed 's/-[0-9]*-g/-/'
fi
