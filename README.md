# OneBot

I've open-sourced this code early, but it's still very much a work in progress. There be dragons ahead.

## Version v0.0.4-wip

Changes happen rapidly, the spec is liquid. Solidified instances will be tagged for go module use.


## Instructions

There's a getting started guide in [GettingStarted.md](./GettingStarted.md).

Build the bot with `make build`.

Run the bot with `make run`.

Build everything and run with `make dev`. Note: This will also trigger `tools/updatelicense.py`, which is a recursive function.

Build the plugins via `make plugins` and `make protocols`. This will build all the plugins in the plugins directory, and all the protocol plugins in the protocols directory.

Check code correctness via `make check`. Note: This tool outputs suggestions, make sure to ask before making changes to already comitted code based on these guidelines.
