module github.com/TheDiscordian/onebot

go 1.18

require (
	github.com/bluesky-social/indigo v0.0.0-20230425215957-9d6296d50061
	github.com/bwmarrin/discordgo v0.20.3
	github.com/lunixbochs/struc v0.0.0-20200707160740-784aaebc1d40
	github.com/matrix-org/gomatrix v0.0.0-20210324163249-be2af5ef2e16
	github.com/pelletier/go-toml v1.9.3
	github.com/syndtr/goleveldb v1.0.0
	go.mongodb.org/mongo-driver v1.7.0
	golang.org/x/text v0.21.0
	mvdan.cc/xurls/v2 v2.5.0
)

require (
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/ipfs/bbloom v0.0.4 // indirect
	github.com/ipfs/go-block-format v0.1.2 // indirect
	github.com/ipfs/go-cid v0.4.0 // indirect
	github.com/ipfs/go-datastore v0.6.0 // indirect
	github.com/ipfs/go-ipfs-blockstore v1.3.0 // indirect
	github.com/ipfs/go-ipfs-ds-help v1.1.0 // indirect
	github.com/ipfs/go-ipfs-util v0.0.2 // indirect
	github.com/ipfs/go-ipld-cbor v0.0.7-0.20230126201833-a73d038d90bc // indirect
	github.com/ipfs/go-ipld-format v0.4.0 // indirect
	github.com/ipfs/go-log v1.0.5 // indirect
	github.com/ipfs/go-log/v2 v2.5.1 // indirect
	github.com/ipfs/go-metrics-interface v0.0.1 // indirect
	github.com/jbenet/goprocess v0.1.4 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multihash v0.2.1 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/polydawn/refmt v0.89.1-0.20221221234430-40501e09de1f // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/whyrusleeping/cbor-gen v0.0.0-20230331140348-1f892b517e70 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	lukechampine.com/blake3 v1.1.7 // indirect
)

replace github.com/TheDiscordian/onebot/libs/onecurrency => ./libs/onecurrency/

replace github.com/TheDiscordian/onebot/onelib => ./onelib/
