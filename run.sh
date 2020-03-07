#! /bin/sh

echo "building icapeg ..."
export GO111MODULE=on
go mod vendor
CGO_ENABLED=0 GOFLAGS=-mod=vendor go build


./icapeg
