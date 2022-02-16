#!/bin/bash

IMG="${IMG:-module-metrics:dev}"

main() {
  docker build -t "${IMG}" .
}

cover() {
  go test -coverprofile=cover.out ./... && \
    go tool cover -html=cover.out -o cover.html
}

"$@"
