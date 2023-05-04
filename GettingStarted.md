# OneBot: Getting Started

This is a quick guide on how to get OneBot running on a Linux PC/VPS.

## Requirements

- A Linux OS
- gcc
- git
- Go (>=v1.11, newer the better)
- Python3

As OneBot doesn't currently have binaries, this guide will tell you how to get it compiled and running. It usually takes under 1min to compile after the first time, and under 5mins the first time. If you need this process to be faster please [open an issue](https://github.com/TheDiscordian/onebot/issues/new), it can be made much faster.

## Getting the Code

First make sure you're not the root user. Then cd to a directory you're comfortable having/running OneBot in, then download the code using git for example:

```bash
cd ~
git clone https://github.com/TheDiscordian/onebot.git
```

## Configuring the Bot

OneBot looks in the current directory for a config file name `onebot.toml` by default. There's an example config located at `onebot/onebot.sample.toml`, copy that file to `onebot/onebot.toml` and edit it:

```bash
cd ~/onebot
cp onebot.sample.toml onebot.toml
```

Open `onebot.toml` using your favourite text editor (if you're very new to shells, try `nano onebot.toml`). You should see the following at the top:

```toml
[general]
default_prefix = ","
default_nickname = "OneBot"
```

You'll want to set these to what you want the command prefix to be, and the default nickname. The command prefix is used for invoking commands for example if it's `,` you could use the `parrot` plugin by saying `,say Hello World!`. The nickname should be whatever you want the bot to be nicknamed (or what it's already nicknamed). This isn't *extremely* important, but if it's wrong some plugins may misbehave.

Now head down to the line defining the protocol plugins, it should look something like this:

```toml
protocols = ["matrix", "discord", "bluesky"]
```

Edit this to only contain protocols you want to use. As of this writing valid values are "matrix", "discord", "bluesky", and "irc_bnetd". In this guide we'll cover "matrix", "discord", and "bluesky" below.

### Matrix

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

### Discord

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

### Bluesky

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

The only required fields are `handle` and `password`. They're pretty self-explanatory, if you login with `discordian.ca`, you use `discordian.ca` as your handle. The password is your password ([app passwords](https://staging.bsky.app/settings/app-passwords) should work too).

OneBot polls the PDS for new posts, and to check if it has any new followers. OneBot only sees posts by people it follows and will automatically follow anyone who follows it. The defaults should be fine for most people, however you're free to adjust these settings as much as your PDS allows.

## Running the Bot

After config running the bot should be pretty straightforward:

```bash
cd ~/onebot
./tools/rundev.sh
```

This will rebuild OneBot and all it's plugins, then run the bot. If you're looking to have it run longer than just your current shell session, install `tmux` (on debian this is `sudo apt install tmux`), ensure the bot isn't running, then try this process:

```bash
tmux
cd ~/onebot
./tools/rundev.sh
CTRL+b, d
```

That last bit is saying "Press CTRL+b, release the keys, then press the d key". This will background a tmux shell, if you want to return to it at any time, ensure you're on your user account and run `tmux attach`.

## Conclusion

Now you should have your own copy of OneBot running! If you run into any problems or things you want please don't hesitate to file a PR or [open an issue](https://github.com/TheDiscordian/onebot/issues/new).