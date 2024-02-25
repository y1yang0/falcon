#! /bin/bash

cd src
go build
rm -rf object
./falcon test/object.y > xx.log 2>&1
./object
