#!/bin/bash
set -euC
set -o xtrace

# curl https://glide.sh/get | sh
glide update

# Only continue if we're on go 1.8; no need to run the linter for every case
if go version | grep -q go1.8; then
    go get -u gopkg.in/alecthomas/gometalinter.v1 && gometalinter.v1 --install
fi
