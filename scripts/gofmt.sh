#!/bin/bash

gofmt_files=$(go fmt ./...)
if [[ -n ${gofmt_files} ]]; then
    echo "needs to run go fmt: ${gofmt_files}"
    exit 1
fi
