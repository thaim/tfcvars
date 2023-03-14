#!/bin/bash

gofmt_files=$(go fmt ./...)
if [[ -n ${gofmt_files} ]]; then
    exit 1
fi
