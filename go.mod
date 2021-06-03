module github.com/TheDiscordian/onebot

go 1.14

require (
	github.com/bwmarrin/discordgo v0.20.3
	github.com/lunixbochs/struc v0.0.0-20200707160740-784aaebc1d40 // indirect
	github.com/matrix-org/gomatrix v0.0.0-20200501121722-e5578b12c752
	github.com/pelletier/go-toml v1.7.0
	github.com/syndtr/goleveldb v1.0.0
	go.mongodb.org/mongo-driver v1.3.3
)

replace github.com/TheDiscordian/gomatrix => ../gomatrix

replace github.com/TheDiscordian/onebot/libs/onecurrency => ./libs/onecurrency/

replace github.com/TheDiscordian/onebot/onelib => ./onelib/
