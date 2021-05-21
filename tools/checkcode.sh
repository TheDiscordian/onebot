#!/bin/sh

echo "Running go vet..."
echo

echo "onebot"
go vet github.com/TheDiscordian/onebot
echo "onelib"
go vet github.com/TheDiscordian/onebot/onelib
echo "./libs/*"
go vet ./libs/*
for f in ./plugins/*.go; do
	echo ${f};
	go vet ${f};
done;
for f in ./protocols/*.go; do
	echo ${f};
	go vet ${f};
done;

if command -v golangci-lint &> /dev/null
then
	echo
	echo "Running golangci-lint..."
	echo

	EXCLUDES='\`Load\` is unused'

	echo "onebot"
	golangci-lint run -e "$EXCLUDES" *.go
	echo "onelib"
	golangci-lint run -e "$EXCLUDES" ./onelib
	echo "./libs/*"
	golangci-lint run -e "$EXCLUDES" ./libs/*
	for f in ./plugins/*.go; do
		echo ${f};
		golangci-lint run -e "$EXCLUDES" ${f};
	done;
	for f in ./protocols/*.go; do
		echo ${f};
		golangci-lint run -e "$EXCLUDES" ${f};
	done;
fi
