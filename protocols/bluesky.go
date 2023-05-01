// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/TheDiscordian/onebot/onelib"
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
)

/* NOTES:
- IDs never contain "$" so for now let's do CID$URI so we have all the data we could possibly need available.
  - It appears BlueSky would like us to send the RootID(s) over too, let's hope that's optional or we'll have to extend this jank...
- ["post"]["author"]["viewer"]["following"] is present if WE follow THEM: "at://did:plc:wjeojdonqywpw5ktsernlecj/app.bsky.graph.follow/3ju2xtrm5zy2l"
- ["post"]["author"]["viewer"]["followedBy"] is present if THEY follow US: "at://did:plc:2cxgdrgtsmrbqnjkwyplmp43/app.bsky.graph.follow/3jqcgeqz74c2q"
- ["post"]["author"]["viewer"]["muted"] is "true" if the user is muted
- ["post"]["record"]["reply"]["root"] contains the "root" thread if available. (If replying "parent" becomes the post you're replying to)
*/

const (
	// NAME is same as filename, minus extension
	NAME = "bluesky"
	// LONGNAME is what's presented to the user
	LONGNAME = "At Protocol"
	// VERSION of the script
	VERSION = "v0.0.0"

	DB_TABLE = "bluesky"
)

var (
	blueskyHandle   string
	blueskyPassword string
	feedCount       int
	feedFreq        int
	followFreq      int
)

func loadConfig() {
	// BlueskyServer = onelib.GetTextConfig(NAME, "server")
	blueskyHandle = onelib.GetTextConfig(NAME, "handle")
	blueskyPassword = onelib.GetTextConfig(NAME, "password")
	feedCount_str := onelib.GetTextConfig(NAME, "feed_count")
	feedCount, _ = strconv.Atoi(feedCount_str)
	feedFreq_str := onelib.GetTextConfig(NAME, "feed_freq")
	feedFreq, _ = strconv.Atoi(feedFreq_str)
	followFreq_str := onelib.GetTextConfig(NAME, "follow_freq")
	followFreq, _ = strconv.Atoi(followFreq_str)
}

// Load connects to Bluesky, and sets up listeners. It's required for OneBot.
func Load() onelib.Protocol {
	loadConfig()
	err := createSession(blueskyHandle, blueskyPassword)
	if err != nil {
		onelib.Error.Println("["+NAME+"] Error creating session:", err)
	}
	bsProto := Bluesky{prefix: onelib.DefaultPrefix, nickname: blueskyHandle, seenPosts: make(map[string]bool), stop: make(chan bool)}
	go bsProto.recv(bsProto.stop)
	go syncFollowers(bsProto.stop)

	return onelib.Protocol(&bsProto)
}

func syncFollowers(stop chan bool) {
	for {
		select {
		case <-stop:
			return
		default:
		}
		follows, err := getFollowsMap()
		if err != nil {
			onelib.Error.Println("["+NAME+"] Error getting follows:", err)
			time.Sleep(time.Duration(followFreq) * time.Second)
			continue
		}
		followers, err := getFollowersMap()
		if err != nil {
			onelib.Error.Println("["+NAME+"] Error getting followers:", err)
			time.Sleep(time.Duration(followFreq) * time.Second)
			continue
		}
		// See who follows us, but we don't follow them and follow them
		for follower := range followers {
			if !follows[follower] {
				err = followUser(follower)
				time.Sleep(time.Second)
				if err != nil {
					onelib.Error.Println("["+NAME+"] Error following user:", err)
				}
			} else {
				delete(follows, follower)
			}
		}
		// See who we follow, but they don't follow us and unfollow them (WIP)
		/*
			for follow := range follows {
				err = unfollowUser(follow)
				if err != nil {
					onelib.Error.Println("["+NAME+"] Error unfollowing user:", err)
				}
			}*/
		time.Sleep(time.Duration(followFreq) * time.Second)
	}
}

func followUser(did string) error {
	auth := getAuthInfo()
	xrpcc, err := getXrpcClient(auth)
	if err != nil {
		return err
	}

	follow := bsky.GraphFollow{
		LexiconTypeID: "app.bsky.graph.follow",
		CreatedAt:     time.Now().Format(time.RFC3339),
		Subject:       did,
	}

	_, err = atproto.RepoCreateRecord(context.TODO(), xrpcc, &atproto.RepoCreateRecord_Input{
		Collection: "app.bsky.graph.follow",
		Repo:       xrpcc.Auth.Did,
		Record:     &lexutil.LexiconTypeDecoder{&follow},
	})
	if err != nil {
		return err
	}

	return nil
}

func unfollowUser(did string) error {
	/*auth := getAuthInfo()
	xrpcc, err := getXrpcClient(auth)
	if err != nil {
		return err
	}

	follow := bsky.GraphFollow{
		LexiconTypeID: "app.bsky.graph.follow",
		CreatedAt:     time.Now().Format(time.RFC3339),
		Subject:       did,
	}

	err := atproto.RepoDeleteRecord(context.TODO(), xrpcc, &atproto.RepoDeleteRecord_Input{
		Collection: "app.bsky.graph.follow",
		Repo:       xrpcc.Auth.Did,
		Record:     &lexutil.LexiconTypeDecoder{&follow},
	})
	if err != nil {
		return err
	}*/
	// Not yet implmented

	return nil
}

func getFollowsMap() (followsMap map[string]bool, err error) {
	auth := getAuthInfo()
	xrpcc, err := getXrpcClient(auth)
	if err != nil {
		return
	}

	follows, err := bsky.GraphGetFollows(context.TODO(), xrpcc, blueskyHandle, "", 100)
	if err != nil {
		return
	}
	followsMap = make(map[string]bool, len(follows.Follows))
	for _, f := range follows.Follows {
		followsMap[f.Did] = true
	}

	return
}

func getFollowersMap() (followersMap map[string]bool, err error) {
	auth := getAuthInfo()
	xrpcc, err := getXrpcClient(auth)
	if err != nil {
		return
	}

	followers, err := bsky.GraphGetFollowers(context.TODO(), xrpcc, blueskyHandle, "", 100)
	if err != nil {
		return
	}
	followersMap = make(map[string]bool, len(followers.Followers))
	for _, f := range followers.Followers {
		followersMap[f.Did] = true
	}

	return
}

func newHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func getXrpcClient(auth *xrpc.AuthInfo) (*xrpc.Client, error) {
	// TODO: Allow custom PDS host
	h := "https://bsky.social"

	return &xrpc.Client{
		Client: newHttpClient(),
		Host:   h,
		Auth:   auth,
	}, nil
}

func getAuthInfo() *xrpc.AuthInfo {
	var auth xrpc.AuthInfo
	jsonAuth, err := onelib.Db.GetString(DB_TABLE, "auth_json")
	if err == nil {
		err := json.Unmarshal([]byte(jsonAuth), &auth)
		if err != nil {
			onelib.Error.Println("["+NAME+"] Error unmarshalling auth_json:", err)
		}
	}
	return &auth
}

func createSession(handle, password string) error {
	auth := getAuthInfo()
	xrpcc, err := getXrpcClient(auth)
	if err != nil {
		return err
	}

	ses, err := atproto.ServerCreateSession(context.TODO(), xrpcc, &atproto.ServerCreateSession_Input{
		Identifier: handle,
		Password:   password,
	})
	if err != nil {
		return err
	}

	b, err := json.Marshal(ses)
	if err != nil {
		return err
	}

	err = onelib.Db.PutString(DB_TABLE, "auth_json", string(b))
	onelib.Debug.Println(string(b))
	return nil
}

func post(text string, reply *bsky.FeedPost_ReplyRef) error {
	auth := getAuthInfo()
	xrpcc, err := getXrpcClient(auth)
	if err != nil {
		return err
	}

	_, err = atproto.RepoCreateRecord(context.TODO(), xrpcc, &atproto.RepoCreateRecord_Input{
		Collection: "app.bsky.feed.post",
		Repo:       auth.Did,
		Record: &lexutil.LexiconTypeDecoder{&bsky.FeedPost{
			Text:      text,
			CreatedAt: time.Now().Format("2006-01-02T15:04:05.000Z"),
			Reply:     reply,
		}},
	})
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}

	//fmt.Println(resp.Cid)
	//fmt.Println(resp.Uri)

	return nil
}

func getFeed(count int64) ([]*bsky.FeedDefs_FeedViewPost, error) {
	auth := getAuthInfo()
	xrpcc, err := getXrpcClient(auth)
	if err != nil {
		return nil, err
	}

	algo := "reverse-chronological"
	tl, err := bsky.FeedGetTimeline(context.TODO(), xrpcc, algo, "", count)
	if err != nil {
		return nil, err
	}

	return tl.Feed, nil
}

// Bluesky is the Protocol object used for handling anything Bluesky related.
type Bluesky struct {
	/*
	   Store useful data here such as connected rooms, admins, nickname, accepted prefixes, etc
	*/
	prefix   string
	nickname string
	stop     chan bool

	seenPosts map[string]bool
}

// Name returns the name of the plugin, usually the filename.
func (bs *Bluesky) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (bs *Bluesky) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (bs *Bluesky) Version() string {
	return VERSION
}

// NewMessage should generate a message object from something
func (bs *Bluesky) NewMessage(raw []byte) onelib.Message {
	// TODO we could construct a message from JSON pretty easily here
	return nil
}

// Send sends a Message object to a location specified by to (usually a location or sender UUID).
func (bs *Bluesky) Send(to onelib.UUID, msg onelib.Message) {
	bs.SendText(to, msg.Text())
}

// SendText sends text to a location specified by to (usually a location or sender UUID).
// However for Bluesky sending text can be to just ... the void. This only supports the void. See
// blueskyLocation for replying to a thread.
func (bs *Bluesky) SendText(to onelib.UUID, text string) {
	post(text, nil)
}

// SendFormattedText sends formatted text to a location specified by to (usually a location or sender UUID).
func (bs *Bluesky) SendFormattedText(to onelib.UUID, text, formattedText string) {
	bs.SendText(to, text)
}

// recv should be called after you've recieved data and built a Message object
func (bs *Bluesky) recv(stop chan bool) {
	var lastCID string
	for {
		select {
		case <-stop:
			return
		default:
		}
		feed, err := getFeed(int64(feedCount))
		if err != nil {
			onelib.Error.Println("["+NAME+"] Error getting feed:", err)
			time.Sleep(time.Duration(feedFreq/2+1) * time.Second)
			createSession(blueskyHandle, blueskyPassword)
			time.Sleep(time.Duration(feedFreq/2+1) * time.Second)
			continue
		}
		firstCID := feed[0].Post.Cid
		for _, item := range feed {
			post := item.Post
			if post.Cid == lastCID {
				break
			}
			if post == nil || post.Author.Handle == blueskyHandle {
				continue
			}
			if bs.seenPosts[post.Cid] {
				break
			}
			bs.seenPosts[post.Cid] = true

			reply := item.Reply

			msg := &bskyMessage{
				bskyPost: &bskyPost{
					cid: post.Cid,
					uri: post.Uri,
				},
				text: post.Record.Val.(*bsky.FeedPost).Text,
			}
			if reply != nil && reply.Root != nil {
				msg.root = &bskyPost{
					cid: reply.Root.Cid,
					uri: reply.Root.Uri,
				}
			}

			location := &bskyLocation{msg: msg}

			sender := &bskySender{
				handle:   post.Author.Handle,
				did:      post.Author.Did,
				location: location,
			}

			onelib.ProcessMessage(bs.prefix, msg, sender)
			onelib.ProcessMessage("@"+blueskyHandle+" ", msg, sender)
		}
		lastCID = firstCID
		time.Sleep(time.Duration(feedFreq) * time.Second)
	}
}

// Remove should disconnect any open connections making it so the bot can forget about the protocol cleanly.
func (bs *Bluesky) Remove() {
	bs.stop <- true
	bs.stop <- true
}

type bskyPost struct {
	cid string
	uri string
}

type bskyMessage struct {
	*bskyPost
	// root of the thread
	root *bskyPost
	// text of the post (may be empty)
	text string
}

func (bm *bskyMessage) UUID() onelib.UUID {
	return onelib.UUID(bm.cid + "$" + bm.uri)
}

func (bm *bskyMessage) Reaction() *onelib.Emoji {
	onelib.Debug.Printf("[%s] Reactions not supported.\n", NAME)
	return nil
}

func (bm *bskyMessage) Text() string {
	return bm.text
}

func (bm *bskyMessage) FormattedText() string {
	return bm.text
}

func (bm *bskyMessage) StripPrefix(prefix string) onelib.Message {
	if len(bm.text) > len(prefix) && strings.HasPrefix(bm.text, prefix) {
		prefix = prefix + " "
	}
	return onelib.Message(&bskyMessage{text: strings.Replace(bm.text, prefix, "", 1)})
}

func (bm *bskyMessage) Raw() []byte {
	// FIXME return the original JSON of the post
	return []byte(bm.text)
}

type bskySender struct {
	handle   string
	location *bskyLocation
	did      string
}

func (bs *bskySender) DisplayName() string {
	return bs.handle
}

func (bs *bskySender) Username() string {
	return bs.handle
}

func (bs *bskySender) UUID() onelib.UUID {
	return onelib.UUID(bs.did)
}

func (bs *bskySender) Location() onelib.Location {
	return bs.location
}

func (bs *bskySender) Protocol() string {
	return NAME
}

func (bs *bskySender) Send(msg onelib.Message) {
	bs.SendText(msg.Text())
}

func (bs *bskySender) SendText(text string) {
	err := post("@"+bs.handle+" "+text, nil)
	if err != nil {
		onelib.Error.Printf("[%s] Error posting message: %s\n", NAME, err)
	}
}

func (bs *bskySender) SendFormattedText(text, formattedText string) {
	// TODO figure out formatted text
	bs.SendText(text)
}

// In Bluesky a message is also a location, a message can be a thread or become a thread.
type bskyLocation struct {
	msg *bskyMessage
}

func (bl *bskyLocation) DisplayName() string {
	return blueskyHandle
}

func (bl *bskyLocation) Nickname() string {
	return blueskyHandle
}

func (bl *bskyLocation) Topic() string {
	return bl.msg.Text()
}

func (bl *bskyLocation) UUID() onelib.UUID {
	return bl.msg.UUID()
}

func (bl *bskyLocation) Send(msg onelib.Message) {
	bl.SendText(bl.msg.Text())
}

func (bl *bskyLocation) SendText(text string) {
	var root *atproto.RepoStrongRef
	if bl.msg.root != nil {
		root = &atproto.RepoStrongRef{Cid: bl.msg.root.cid, Uri: bl.msg.root.uri}
	} else {
		root = &atproto.RepoStrongRef{Cid: bl.msg.cid, Uri: bl.msg.uri}
	}
	post(text,
		&bsky.FeedPost_ReplyRef{
			Parent: &atproto.RepoStrongRef{Cid: bl.msg.cid, Uri: bl.msg.uri},
			Root:   root,
		})
}

func (bl *bskyLocation) SendFormattedText(text, formattedText string) {
	// TODO: Proper formatted text
	bl.SendText(text)
}

func (bl *bskyLocation) Protocol() string {
	return NAME
}
