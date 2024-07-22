#!/bin/bash

# prints out current version from CI variables

set -euo pipefail

echoerr() { printf "%s\n" "$*" >&2; }
if [[ ${CI_COMMIT_TAG:-} != "" ]]; then
  echo "latest"
elif [[ ${CI_COMMIT_BRANCH:-} != "" ]]; then
  echo $CI_COMMIT_BRANCH | sed 's/\//-/'
else
  echoerr "CI_COMMIT_TAG and CI_COMMIT_BRANCH undefined, can't get branch"
  exit 1
fi
