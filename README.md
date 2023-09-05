# OneBot
[![Onelib Go Reference](https://pkg.go.dev/badge/github.com/TheDiscordian/onebot/onelib.svg)](https://pkg.go.dev/github.com/TheDiscordian/onebot/onelib)

OneBot is a multi-protocol bot driven by feature plugins. It's primary purpose is for chatting, but it's flexible enough to be used for a variety of things.

## Version v0.1.0

### Features

OneBot is powered by plugins, this will be a list of features and what plugins provide them.

#### Protocols

OneBot can connect to, read, and send messages on the following protocols:

- [AtProtocol](https://atproto.com/) ([Bluesky](https://bsky.app/)) ([bluesky.go](protocols/bluesky.go))
- [Discord](https://discord.com) ([discord.go](protocols/discord.go))
- [Bnetd IRC Chat](https://pvpgn.fandom.com/wiki/BNETD#IRC_Settings) ([irc_bnetd.go](protocols/irc_bnetd.go))
- [Matrix](https://matrix.org/) ([matrix.go](protocols/matrix.go))

#### Plugins

OneBot supports the following plugins and features. This section will also list commands provided by the plugins. A command for `parrot` for example is `say`, and the default command prefix is `,`. So if a user typed `,say Hello World!` the bot would reply with `Hello World!`.

- 8Ball ([8ball.go](plugins/8ball.go))
	- `8ball <question>` / `8b <question>`
		- Predicts the future.
- Bash Quotes ([bashquotes.go](plugins/bashquotes.go))
	- `bash`
		- Gets a random quote from [bash.org](https://bash.org/), and shares it.
- BNetd Bridge ([bnetdbridge.go](plugins/bnetdbridge.go))
	- Bridges a BNetd IRC channel with another protocol.
- XKCD Comics ([comics.go](plugins/comics.go))
	- `comic`
		- Returns a random XKCD comic complete with link and flavour text if possible.
- Diablo II Leaderboards ([d2lb.go](plugins/d2lb.go))
	- Plugin description:
		- This plugin is meant to be used on a Diablo II BNetd server. It can read the `ladder.D2DV` file to display formatted leaderboards for various difficulties.
	- `d2xp [count]`
		- Returns the Diablo II Expansion leaderboard up to count.
	- `d2 [count]`
		- Returns the Diablo II Standard leaderboard up to count.
	- `d2hc [count]`
		- Returns the Diablo II Hardcore leaderboard up to count.
	- `d2xphc [count]`
		- Returns the Diablo II Expansion Hardcore leaderboard up to count.
	- `d2all [count]`
		- Returns all the leaderboards up to count.
- Dice ([dice.go](plugins/dice.go))
	- `roll [sides]` / `r [sides]`
		- Rolls one die.
- Fact Crow ([factcrow.go](plugins/factcrow.go))
	- `fc` / `fact`
		- Returns a random "fact".
- IPFS ([ipfs.go](plugins/ipfs.go))
	- `ipfs-check <multiaddr> <CID> [Backend URL]`
		- Check information about a CID and multiaddress. For more information see [ipfs-check.on.fleek.co](https://ipfs-check.on.fleek.co/).
	- `ipfs-findprovs <CID>`
		- Runs `ipfs dht findprovs <CID>` on a local Kubo node, returning the result if possible.
	- `ipfs-stat <CID>`
		- Runs `ipfs block stat <CID>` on a local Kubo node, returning the result if possible.
	- `web3-help`
		- Returns help text about web3-store.
	- `web3-store <cid>`
		- Store up to 100MB on web3.storage.
- Currency ([money.go](plugins/money.go))
	- Plugin description:
		- A currency and leaderboard system. This plugin adds features for users to gain a virtual currency, see each-other on leaderboards, and deposit/withdraw the currency. It also has some old, possibly unstable aliasing features to help link IDs together.
	- `bal [ID]` / `balance [ID]`
		- Check your current balance, or optionally check someone else's balance.
	- `cute`
		- Attempt to gain currency by doing something cute.
	- `chill`
		- Attempt to gain currency by doing something chill.
	- `meme`
		- Attempt to gain currency by doing something memey.
	- `risk`
		- Attempt to gain currency by doing something risky.
	- `dep <amount / all>` / `deposit <amount / all>`
		- Deposits currency into the bank.
	- `withdraw <amount / all>`
		- Withdraws currency from the bank.
	- `alias <UID>`
		- Turns current user into target of another user.
	- `unalias`
		- Removes alias on current user.
	- `confirmalias <UID>`
		- Confirms an alias.
	- `leaderboard` / `lb`
		- See the leaderboard of who has the most currency.
- Parrot ([parrot.go](plugins/parrot.go))
	- `say [text]` / `s [text]`
		- Repeats the text back to the sender.
	- `rev [text]` / `r [text]`
		- Repeats the text back to the sender in reverse.
- Question & Answer ([qa.go](plugins/qa.go))
	- `q <question>` / `question <question>`
		- Uses OpenAI and embeds to answer the question.
	- `stats [yyyy-mm]`
		- Returns approval rating of questions within a certain timeframe.
- Role Triggers (Discord only) ([roletriggers.go](plugins/roletriggers.go))
	- Plugin description:
		- This plugin can be configured by an admin to add / remove roles from users based on how the react to specific messages. The Discord admin can be specified under the Discord protocol plugin's configuration.
	- `roleid <role>`
		- Returns the roleid of a role.
	- `addtrigger <messageID> <emoji> <@role>` / `at <messageID> <emoji> <@role>`
		- Adds a trigger to a message which will give any user who reacts with the specified emoji, the specified role (and remove said role upon removing the reaction).
	- `removetrigger <messageID> <emoji>` / `rt <messageID> <emoji>`
		- Removes specified emoji trigger from specified message.

## Getting Started

### Requirements

- A Linux OS
	- Will work with MacOS as well, but you'd need to build your own binaries.

### Getting OneBot

In your terminal run the follow commands as an unprivileged user in their home folder:

```bash
wget https://github.com/TheDiscordian/onebot/releases/latest/download/onebot-linux64.tar.xz
tar -xf onebot-linux64.tar.xz
mv bin/release ./onebot
rmdir bin
```

OneBot will now be located in the `onebot` directory.

## Configuring OneBot

OneBot looks in the current directory for a config file name `onebot.toml` by default. Open `onebot.toml` using your favourite text editor (if you're very new to shells, try `nano onebot.toml`). You should see the following at the top:

```toml
[general]
default_prefix = ","
default_nickname = "OneBot"
```

You'll want to set these to what you want the command prefix to be, and the default nickname. The command prefix is used for invoking commands for example if it's `,` you could use the `parrot` plugin by saying `,say Hello World!`. The nickname should be whatever you want the bot to be nicknamed (or what it's already nicknamed). This isn't *extremely* important, but if it's wrong some plugins may misbehave.

### Protocols

In `onebot.toml`, head down to the line defining the protocol plugins, it should look something like this:

```toml
protocols = ["matrix", "discord", "bluesky"]
```

Edit this to only contain protocols you want to use.

#### Matrix

The Matrix config section looks like this:

```toml
[matrix]
# for example: https://matrix.org
home_server = ""
# if using auth_token, it must be in "@username:homeserver" format
auth_user = ""
# if blank, falls back onto auth_pass
auth_token = ""
# if set, this will be used to retrieve an auth token, then the password can be omitted from this file
auth_pass = ""
```

It's pretty self-explanatory. You set the homeserver of the bot, just like you would on a Matrix client. You set your username also as you normally would. If you have an `auth_token`, you can set that here too. If you don't have one, you can set `auth_pass` instead. Once the bot successfully logs in once it will store an `auth_token` in it's database, so you can remove the password from this file after.

#### Discord

The Discord config section looks like this:

```toml
[discord]
# is set as "Bot auth_token"
auth_token = ""
# id of admin user (for certain plugins like 'roletriggers')
admin_id = ""
```

You can get an `auth_token` for the bot from the [Discord Developer Portal](https://discord.com/developers/applications) (if you don't know how, check out [this guide](https://discordgsm.com/guide/how-to-get-a-discord-bot-token)).

`admin_id` isn't required, but can be helpful for plugins like `roletriggers`. ID here would be a Discord UID of the admin user.

#### Bluesky

The Bluesky protocol is pretty new, and only supports `https://bsky.social` as a PDS for now, but this is expected to change in the near future. It's section looks like this:

```toml
[bluesky]
# The handle on Bluesky of the account to use
handle = ""
# The password of the Bluesky account
password = ""
# The number of posts to fetch at once per poll
feed_count = 50
# The number of seconds to wait in between polls
feed_freq = 10
# The number of seconds to wait in between syncing followers
follow_freq = 60
```

The only required fields are `handle` and `password`. They're pretty self-explanatory, if you login with `discordian.ca`, you use `discordian.ca` as your handle. The password is your password ([app passwords](https://staging.bsky.app/settings/app-passwords) work too).

OneBot polls the PDS for new posts, and to check if it has any new followers. OneBot only sees posts by people it follows and will automatically follow anyone who follows it. The defaults should be fine for most people, however you're free to adjust these settings as much as your PDS allows.

### Running OneBot

The simplest way to run OneBot after configuration is simply to run the binary:

```bash
cd ~/onebot
./onebot
```

However if you'd like OneBot to persist outside of just this terminal session, then I highly recommend using `tmux`. Install tmux, ensure OneBot isn't running, then try the following:

```bash
tmux
cd ~/onebot
./onebot # run the bot
CTRL+b, d
```

That last bit is saying "Press CTRL+b, release the keys, then press the d key". This will background a tmux shell, if you want to return to it at any time, ensure you're on your user account and run `tmux attach`.

## Building OneBot

### Requirements

- Linux, MacOS, or a BSD
- gcc
- git
- make
- Go (>=v1.20)
- Python3 (optional)

### Instructions

Get the source code with `git clone https://github.com/TheDiscordian/onebot.git`

Build the bot with `make build`.

Run the bot with `make run`.

Build everything and run with `make dev`. Note: This will also trigger `tools/updatelicense.py`, which is a recursive function.

Build the plugins via `make plugins` and `make protocols`. This will build all the plugins in the plugins directory, and all the protocol plugins in the protocols directory.

Check code correctness via `make check`. Note: This tool outputs suggestions, make sure to ask before making changes to already comitted code based on these guidelines.