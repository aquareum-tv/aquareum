#!/bin/bash

set -euo pipefail
set -x

cd /c/Users/admin/AppData/Local/aquareum_desktop
cd app-*
./Aquareum.exe -- --self-test
