.PHONY: plugins protocols

dev: license build run

license:
	 python3 ./tools/checklicense.py

plugins:
	./tools/build_plugins.sh

protocols:
	./tools/build_protocols.sh

build: plugins protocols
	go build -ldflags "-s -w" -trimpath -o bin/onebot *.go

run:
	./bin/onebot

check:
	./tools/checkcode.sh