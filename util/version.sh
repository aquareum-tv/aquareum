#!/bin/bash

# prints out current version from git tags

set -euo pipefail

git describe --long --tags --dirty | sed 's/-[0-9]*-g/-/'
