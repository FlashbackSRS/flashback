#!/bin/bash
set -euC
set -o xtrace

if [ "$TRAVIS_OS_NAME" == "osx" ]; then
    brew install glide
fi

glide update

# Only run GopherJS tests and Linter tests on Linux/Go 1.9
if [[ ${TRAVIS_OS_NAME} == "linux" && ${TRAVIS_GO_VERSION} == 1.9* ]]; then
    if [ "$TRAVIS_OS_NAME" == "linux" ]; then
        # Install nodejs and dependencies, but only for Linux
        curl -sL https://deb.nodesource.com/setup_6.x | sudo -E bash -
        sudo apt-get update -qq
        sudo apt-get install -y nodejs
    fi
    npm install
    # Install Go deps only needed by PouchDB driver/GopherJS
    [ -e glide.gopherjs.yaml ] && glide -y glide.gopherjs.yaml install
    # Then install GopherJS and related dependencies
    go get -u github.com/gopherjs/gopherjs

    # Source maps (mainly to make GopherJS quieter; I don't really care
    # about source maps in Travis)
    npm install source-map-support

    # Set up GopherJS for syscalls
    (
        cd $GOPATH/src/github.com/gopherjs/gopherjs/node-syscall/
        npm install --global node-gyp
        node-gyp rebuild
        mkdir -p ~/.node_libraries/
        cp build/Release/syscall.node ~/.node_libraries/syscall.node
    )

    go get -u -d -tags=js github.com/gopherjs/jsbuiltin

    # Linter
    go get -u gopkg.in/alecthomas/gometalinter.v1 && gometalinter.v1 --install
fi
