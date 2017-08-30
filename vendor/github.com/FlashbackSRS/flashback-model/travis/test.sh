#!/bin/bash
set -euC

diff -u <(echo -n) <(gofmt -e -d $(find . -type f -name '*.go' -not -path "./vendor/*"))
gometalinter.v1 --config .linter_test.json
gometalinter.v1 --config .linter.json

echo "" >| coverage.txt

for d in $(go list ./... | grep -v /vendor/); do
    gopherjs test $d
    go test -race -coverprofile=profile.out -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done

if [ "${CI:-}" == "true" ]; then
    bash <(curl -s https://codecov.io/bash)
fi
