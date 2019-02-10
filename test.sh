#!/usr/bin/env bash

set -e
export GO111MODULE=on
echo "" > coverage.txt

for d in $(go list ./... | grep -v vendor); do
    go test -race -coverprofile=profile.out -covermode=atomic ${d}
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done

if [ -f "./codecov.sh" ]; then
   ./codecov.sh
fi
