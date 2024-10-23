#!/bin/bash

set -euo pipefail
set -x

curl -L -o aquareum-desktop.exe "$1"
./aquareum-desktop.exe
powershell -Command "Stop-Process -Name Aquareum"
cd /c/Users/admin/AppData/Local/aquareum_desktop
cd app-*
./Aquareum.exe -- --self-test
