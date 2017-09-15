#!/bin/bash
set -euC
set -o xtrace

if [ "$TRAVIS_OS_NAME" == "osx" ]; then
    brew install glide
fi

glide install

function generate {
    go get -u github.com/jteeuwen/go-bindata/...
    go generate $(go list ./... | grep -v /vendor/)
}

function wait_for_server {
    printf "Waiting for $1"
    n=0
    until [ $n -gt 5 ]; do
        curl --output /dev/null --silent --head --fail $1 && break
        printf '.'
        n=$[$n+1]
        sleep 1
    done
    printf "ready!\n"
}

function setup_couch16 {
    if [ "$TRAVIS_OS_NAME" == "osx" ]; then
        return
    fi
    docker pull couchdb:1.6.1
    docker run -d -p 6000:5984 -e COUCHDB_USER=admin -e COUCHDB_PASSWORD=abc123 --name couchdb16 couchdb:1.6.1
    wait_for_server http://localhost:6000/
    curl --silent --fail -o /dev/null -X PUT http://admin:abc123@localhost:6000/_config/replicator/connection_timeout -d '"5000"'
}

function setup_couch20 {
    if [ "$TRAVIS_OS_NAME" == "osx" ]; then
        return
    fi
    docker pull klaemo/couchdb:latest
    docker run -d -p 6001:5984 -e COUCHDB_USER=admin -e COUCHDB_PASSWORD=abc123 --name couchdb20 klaemo/couchdb:2.0.0
    wait_for_server http://localhost:6001/
    curl --silent --fail -o /dev/null -X PUT http://admin:abc123@localhost:6001/_users
    curl --silent --fail -o /dev/null -X PUT http://admin:abc123@localhost:6001/_replicator
    curl --silent --fail -o /dev/null -X PUT http://admin:abc123@localhost:6001/_global_changes
}

case "$1" in
    "standard")
        setup_couch16
        setup_couch20
        generate
    ;;
    "linter")
        go get -u gopkg.in/alecthomas/gometalinter.v1
        gometalinter.v1 --install
    ;;
    "coverage")
        generate
    ;;
esac
