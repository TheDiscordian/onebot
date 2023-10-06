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

rel: build
	mkdir -p bin/release/plugins
	mkdir -p bin/release/protocols
	cp onebot.sample.toml bin/release/onebot.sample.toml
	mv bin/onebot bin/release/onebot
	mv plugins/*.so bin/release/plugins
	cp -R plugins/qa bin/release/plugins
	mv protocols/*.so bin/release/protocols
	cp LICENSE bin/release/LICENSE
	cp README.md bin/release/README.md
	cp CONTRIBUTORS bin/release/CONTRIBUTORS
	tar -caf bin/onebot-linux64.tar.xz bin/release

run:
	./bin/onebot

check:
	./tools/checkcode.sh