#!/bin/bash -e

go-fuzz-build

rm -f crashers/*
cp manualcorpus/* corpus/
exec nice go-fuzz -bin gofuzzdep-fuzz.zip -timeout 1
