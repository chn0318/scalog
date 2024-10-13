#!/bin/bash
set -e
echo "---------- Updating package list ----------"
sudo apt update
echo "---------- Installing Go ----------"
cd /usr/local 
sudo wget https://go.dev/dl/go1.23.1.linux-arm64.tar.gz
sudo tar -C /usr/local -xzf go1.23.1.linux-arm64.tar.gz
export PATH=$PATH:/usr/local/go/bin
export GOPATH=~/Project-Go
export PATH=$PATH:$GOPATH/bin
echo "---------- Installing goreman ----------"
go install github.com/mattn/goreman@latest
cd ~
echo "---------- Downloading and Complie scalog ----------"
mkdir -p ./Project-Go/src/github.com/scalog
cd ./Project-Go/src/github.com/scalog
git clone https://github.com/chn0318/scalog.git
cd ./scalog
export GO111MODULE=off
go build
echo "---------- Adding Env var ----------"
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=~/Project-Go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc