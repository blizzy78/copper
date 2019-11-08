#!/bin/bash

set -e

go install \
	github.com/dvyukov/go-fuzz/go-fuzz \
	github.com/dvyukov/go-fuzz/go-fuzz-build

./fuzz-gopath.sh
export GOPATH=$PWD/gopath
export GO111MODULES=off

go-fuzz-build

cp manualcorpus/* corpus/

nice go-fuzz -bin gofuzzdep-fuzz.zip -timeout 1
