#!/bin/bash

# disable cgo
export CGO_ENABLED=0

set -e
set -x

go build -o release/linux/amd64/plugin