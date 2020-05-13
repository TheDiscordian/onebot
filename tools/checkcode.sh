#!/bin/sh

go vet github.com/TheDiscordian/onebot
go vet github.com/TheDiscordian/onebot/onelib
# TODO vet plugins


# Using relative paths, see: https://github.com/golang/lint/issues/409
golint
golint ./onelib
golint ./plugins/*.go
golint ./protocols/*.go
