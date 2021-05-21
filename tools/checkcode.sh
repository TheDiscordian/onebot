#!/bin/sh

go vet github.com/TheDiscordian/onebot
go vet github.com/TheDiscordian/onebot/onelib
go vet ./libs/*
for f in ./plugins/*.go; do
	echo ${f};
	go vet ${f};
done;
for f in ./protocols/*.go; do
	echo ${f};
	go vet ${f};
done;
