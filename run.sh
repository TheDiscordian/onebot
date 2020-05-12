#!/bin/sh
go fmt ./...
./buildplugins.sh
go build -ldflags "-s -w" -trimpath -o bin/onebot *.go
./bin/onebot
