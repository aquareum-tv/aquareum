#!/bin/bash

set -euo pipefail
set -x

curl -L -o aquareum-desktop.exe https://git.aquareum.tv/aquareum-tv/aquareum/-/package_files/2844/download
./aquareum-desktop.exe
cd /c/Users/admin/AppData/Local/aquareum_desktop
cd app-*
./Aquareum.exe -- --self-test
