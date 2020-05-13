#!/bin/bash

plugins=$(ls ./plugins | grep .go)
protocols=$(ls ./protocols | grep .go)

echo "Building plugins..."
while IFS= read -r line; do
	echo "$line"
	go build -buildmode=plugin -ldflags "-s -w" -trimpath -o ./plugins ./plugins/$line
done <<< "$plugins"

echo "Building protocols..."
while IFS= read -r line; do
	echo "$line"
	go build -buildmode=plugin -ldflags "-s -w" -trimpath -o ./protocols ./protocols/$line
done <<< "$protocols"
