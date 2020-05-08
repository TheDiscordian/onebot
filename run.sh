#!/bin/sh
go fmt ./...
./buildplugins.sh
go run -ldflags "-s -w" -trimpath *.go
