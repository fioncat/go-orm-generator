#!/bin/bash
VERSION=$1
FILE1=go-gendb_$(echo $VERSION)_darwin.zip
FILE2=go-gendb_$(echo $VERSION)_linux.zip

go build
zip -qrm -o $FILE1 go-gendb

GOOS=linux GOARCH=amd64 go build
zip -qrm -o $FILE2 go-gendb


