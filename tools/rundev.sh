#!/bin/sh
go fmt ./...
python ./tools/checklicense.py
./tools/buildplugins.sh
go build -ldflags "-s -w" -trimpath -o bin/onebot *.go
./bin/onebot
