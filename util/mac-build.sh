#!/bin/bash

set -euo pipefail
set -x

# setup environment
echo 'admin' | sudo -S umount "/Volumes/My Shared Files"
mkdir ~/build
mount_virtiofs com.apple.virtio-fs.automount ~/build
eval "$(/opt/homebrew/bin/brew shellenv)"
brew update --force --quiet
cd ~/build/aquareum

brew install ninja go openssl@3 && go version
sudo gem install --user-install xcpretty
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs > rustup.sh && bash rustup.sh -y && rm rustup.sh
export PATH="$PATH:$HOME/.cargo/bin:$(find $HOME/.gem/ruby -type d -name bin -maxdepth 2)"
brew install python@3.11 node
python3.11 -m pip install virtualenv
python3.11 -m virtualenv ~/venv
source ~/venv/bin/activate
pip3 install meson
make node-all-platforms-macos -j16
