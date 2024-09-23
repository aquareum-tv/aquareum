#!/bin/bash

set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
DIR="$(realpath $SCRIPT_DIR/..)"

tart list --format json | jq -r '.[] | select(.State == "running") | .Name' | xargs -L 1 tart stop
IMAGE=sonoma-$(date +%s)
tart clone ghcr.io/cirruslabs/macos-sonoma-base:latest $IMAGE
bash -c "tart run $IMAGE --no-graphics --dir=aquareum:$DIR &"
while ! tart ip $IMAGE; do echo 'waiting for ip...' && sleep 1; done;
cat util/mac-build.sh | sshpass -p admin ssh -o "StrictHostKeyChecking no" admin@$(tart ip $IMAGE) bash -c 'cat > mac-build.sh && bash mac-build.sh' || echo 'build failed'
tart stop $IMAGE
