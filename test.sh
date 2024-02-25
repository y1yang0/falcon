#! /bin/bash

cd src/test
go clean -testcache
go test ./...
