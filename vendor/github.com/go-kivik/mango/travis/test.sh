#!/bin/bash
set -euC
set -o xtrace

go test -race -cover -covermode=atomic -coverprofile=coverage.txt

# Only continue if we're on go 1.8; no need to run the linter for every case
if go version | grep -q go1.8; then
    diff -u <(echo -n) <(gofmt -e -d $(find . -type f -name '*.go' -not -path "./vendor/*"))
    gometalinter.v1 --config .linter_test.json
    gometalinter.v1 --config .linter.json
    bash <(curl -s https://codecov.io/bash)
fi
