# OneBot
[![Onelib Go Reference](https://pkg.go.dev/badge/github.com/TheDiscordian/onebot/onelib.svg)](https://pkg.go.dev/github.com/TheDiscordian/onebot/onelib)

OneBot is a multi-protocol bot driven by feature plugins. It's primary purpose is for chatting, but it's flexible enough to be used for a variety of things.

## Version v0.0.5-wip

Changes happen rapidly, the spec is liquid. Solidified instances will be tagged for go module use.


## Instructions

There's a getting started guide in [GettingStarted.md](./GettingStarted.md).

Build the bot with `make build`.

Run the bot with `make run`.

Build everything and run with `make dev`. Note: This will also trigger `tools/updatelicense.py`, which is a recursive function.

Build the plugins via `make plugins` and `make protocols`. This will build all the plugins in the plugins directory, and all the protocol plugins in the protocols directory.

Check code correctness via `make check`. Note: This tool outputs suggestions, make sure to ask before making changes to already comitted code based on these guidelines.
