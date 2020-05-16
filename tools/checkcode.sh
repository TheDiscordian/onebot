#!/bin/sh

go vet github.com/TheDiscordian/onebot
go vet github.com/TheDiscordian/onebot/onelib
go vet ./libs/*
# TODO vet plugins


# Using relative paths, see: https://github.com/golang/lint/issues/409
golint
golint ./onelib
golint ./libs/*
golint ./plugins/*.go
golint ./protocols/*.go
