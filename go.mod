module github.com/TheDiscordian/onebot

go 1.14

require (
	github.com/bluesky-social/indigo v0.0.0-20230425215957-9d6296d50061
	github.com/bwmarrin/discordgo v0.20.3
	github.com/lunixbochs/struc v0.0.0-20200707160740-784aaebc1d40
	github.com/matrix-org/gomatrix v0.0.0-20210324163249-be2af5ef2e16
	github.com/pelletier/go-toml v1.9.3
	github.com/syndtr/goleveldb v1.0.0
	go.mongodb.org/mongo-driver v1.7.0
	golang.org/x/text v0.8.0
	mvdan.cc/xurls/v2 v2.5.0
)

replace github.com/TheDiscordian/onebot/libs/onecurrency => ./libs/onecurrency/

replace github.com/TheDiscordian/onebot/onelib => ./onelib/
