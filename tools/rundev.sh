#!/bin/sh
go fmt ./...
python3 ./tools/checklicense.py
./tools/buildplugins.sh
go build -ldflags "-s -w" -trimpath -o bin/onebot *.go
./bin/onebot
