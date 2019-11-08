#!/bin/bash

set -e

# hacky way to make go-fuzz work with modules.
go mod vendor
rm -rf gopath
mkdir -p gopath/src/
mv vendor/* gopath/src/
rm -rf vendor
