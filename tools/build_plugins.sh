#!/bin/bash

plugins=$(ls ./plugins | grep .go)

while IFS= read -r line; do
	echo "$line"
	go build -buildmode=plugin -ldflags "-s -w" -trimpath -o ./plugins ./plugins/$line
done <<< "$plugins"
