#!/bin/bash
set -euC

if [ "${TRAVIS_OS_NAME:-}" == "osx" ]; then
    # We don't have docker in OSX, so skip these tests
    unset KIVIK_TEST_DSN_COUCH16
    unset KIVIK_TEST_DSN_COUCH20
fi

function join_list {
    local IFS=","
    echo "$*"
}

go test -race $(go list ./... | grep -v /vendor/)

# Only run GopherJS tests,  Linter and coveragetests on Linux/Go 1.9
if [[ ${TRAVIS_OS_NAME} == "linux" && ${TRAVIS_GO_VERSION} == 1.9* ]]; then
    gopherjs test $(go list ./... | grep -v /vendor/)

    # Linter
    diff -u <(echo -n) <(gofmt -e -d $(find . -type f -name '*.go' -not -path "./vendor/*"))
    go install # to make gotype (run by gometalinter) happy
    gometalinter.v1 --config .linter_test.json
    gometalinter.v1 --config .linter.json

    # Coverage
    go test -coverprofile=coverage.txt -covermode=set && bash <(curl -s https://codecov.io/bash)
fi
