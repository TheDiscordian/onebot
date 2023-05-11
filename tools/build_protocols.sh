#!/bin/bash

protocols=$(ls ./protocols | grep .go)

while IFS= read -r line; do
	echo "$line"
	go build -buildmode=plugin -ldflags "-s -w" -trimpath -o ./protocols ./protocols/$line
done <<< "$protocols"
