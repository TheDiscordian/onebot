// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package main

import (
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/TheDiscordian/onebot/onelib"
	"github.com/matrix-org/gomatrix"
)

const (
	// NAME is same as filename, minus extension
	NAME = "matrix"
	// LONGNAME is what's presented to the user
	LONGNAME = "Matrix"
	// VERSION of the script
	VERSION = "v0.0.0"
)

var (
	// matrixHomeServer
	matrixHomeServer string
	// matrixAuthUser
	matrixAuthUser string
	// matrixAuthToken if blank, falls back onto pass
	matrixAuthToken string
	// matrixAuthPass
	matrixAuthPass string
)

func loadConfig() {
	matrixHomeServer = onelib.GetTextConfig(NAME, "home_server")
	matrixAuthUser = onelib.GetTextConfig(NAME, "auth_user")
	matrixAuthToken = onelib.GetTextConfig(NAME, "auth_token")
	matrixAuthPass = onelib.GetTextConfig(NAME, "auth_pass")
}

// Matrix protocol structs. These should maybe be in their own library.

// RespUserStatus is the JSON response for https://matrix.org/docs/spec/client_server/r0.6.0#get-matrix-client-r0-presence-userid-status
// https://github.com/matrix-org/gomatrix/pull/80
type RespUserStatus struct {
	Presence        string `json:"presence"`
	StatusMsg       string `json:"status_msg"`
	lastActiveAgo   int    `json:"last_active_ago"`
	currentlyActive bool   `json:"currently_active"`
}

// GetStatus returns the status of the user from the specified MXID. See https://matrix.org/docs/spec/client_server/r0.6.0#get-matrix-client-r0-presence-userid-status
// https://github.com/matrix-org/gomatrix/pull/80
func (cli *matrixClient) GetStatus(mxid string) (resp *RespUserStatus, err error) {
	urlPath := cli.BuildURL("presence", mxid, "status")
	err = cli.MakeRequest("GET", urlPath, nil, &resp)
	return
}

// GetOwnStatus returns the user's status. See https://matrix.org/docs/spec/client_server/r0.6.0#get-matrix-client-r0-presence-userid-status
// https://github.com/matrix-org/gomatrix/pull/80
func (cli *matrixClient) GetOwnStatus() (resp *RespUserStatus, err error) {
	return cli.GetStatus(cli.UserID)
}

// SetStatus sets the user's status. See https://matrix.org/docs/spec/client_server/r0.6.0#put-matrix-client-r0-presence-userid-status
// https://github.com/matrix-org/gomatrix/pull/80
func (cli *matrixClient) SetStatus(presence, status string) (err error) {
	urlPath := cli.BuildURL("presence", cli.UserID, "status")
	s := struct {
		Presence  string `json:"presence"`
		StatusMsg string `json:"status_msg"`
	}{presence, status}
	err = cli.MakeRequest("PUT", urlPath, &s, nil)
	return
}

// MarkRead marks eventID in roomID as read, signifying the event, and all before it have been read. See https://matrix.org/docs/spec/client_server/r0.6.0#post-matrix-client-r0-rooms-roomid-receipt-receipttype-eventid
// https://github.com/matrix-org/gomatrix/pull/81
func (cli *matrixClient) MarkRead(roomID, eventID string) error {
	urlPath := cli.BuildURL("rooms", roomID, "receipt", "m.read", eventID)
	return cli.MakeRequest("POST", urlPath, nil, nil)
}

type matrixProtocolMessage struct {
	Format        string `json:"format"`
	Msgtype       string `json:"msgtype"`
	Body          string `json:"body"`
	FormattedBody string `json:"formatted_body"`
}

func (client *matrixClient) setAvatarToFile(fPath string) error {
	if onelib.DefaultAvatar == "" {
		return errors.New("no default avatar set")
	}
	f, err := os.Open(fPath)
	if err != nil {
		return err
	}
	fInfo, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	var resp *gomatrix.RespMediaUpload
	resp, err = client.UploadToContentRepo(f, "image/png", fInfo.Size()) // TODO don't assume png
	if err != nil {
		f.Close()
		return err
	}
	onelib.Info.Println("Avatar set! ContentURI:", resp.ContentURI)
	client.SetAvatarURL(resp.ContentURI)
	f.Close()
	return nil
}

// Load connects to Matrix, and sets up listeners. It's required for OneBot.
// TODO store rooms as a map of locations mapped by UID
func Load() onelib.Protocol {
	loadConfig()

	client, err := gomatrix.NewClient(matrixHomeServer, matrixAuthUser, matrixAuthToken)
	if err != nil {
		onelib.Error.Panicln(err)
	}
	if matrixAuthToken == "" {
		if matrixAuthUser == "" {
			panic("both matrixAuthPass and matrixAuthToken can't be blank.")
		}
		resp, err := client.Login(&gomatrix.ReqLogin{
			Type:     "m.login.password",
			User:     matrixAuthUser,
			Password: matrixAuthPass,
		})
		if err != nil {
			onelib.Error.Panicln(err)
		}
		onelib.SetTextConfig(NAME, "auth_token", resp.AccessToken)
		onelib.Info.Println("Access token (saved):", resp.AccessToken)
		client.SetCredentials(resp.UserID, resp.AccessToken)
	}
	syncer := client.Syncer.(*gomatrix.DefaultSyncer)

	matrix := &Matrix{client: &matrixClient{Client: client}, prefix: onelib.DefaultPrefix, nickname: onelib.DefaultNickname, knownMembers: new(memberMap)}
	matrix.knownMembers.mMap = make(map[onelib.UUID]*member, 1)
	matrix.knownMembers.lock = new(sync.RWMutex)

	syncer.OnEventType("m.room.message", func(ev *gomatrix.Event) {
		if ev.Content["msgtype"] != nil && ev.Content["msgtype"].(string) == "m.text" {
			msg := &matrixMessage{text: ev.Content["body"].(string)} // FIXME just because msgtype is m.text, doesn't mean body *absolutely* exists
			if ev.Content["format"] != nil && ev.Content["format"].(string) == "org.matrix.custom.html" {
				msg.formattedText = ev.Content["formatted_body"].(string) // FIXME just because format is org.matrix.custom.html, doesn't mean formatted_body *absolutely* exists
			}
			mc := &matrixClient{Client: client}
			ml := &matrixLocation{Client: mc, uuid: onelib.UUID(ev.RoomID)}
			var displayName string
			if matrix.knownMembers.Get(onelib.UUID(ev.Sender)) == nil {
				resp, err := client.GetDisplayName(ev.Sender)
				if err != nil {
					displayName = ev.Sender
				}
				displayName = resp.DisplayName
				matrix.knownMembers.Set(onelib.UUID(ev.Sender), &member{displayName: resp.DisplayName})
			} else {
				displayName = matrix.knownMembers.Get(onelib.UUID(ev.Sender)).displayName
			}
			sender := &matrixSender{uuid: onelib.UUID(ev.Sender), username: ev.Sender, displayName: displayName, location: ml}
			matrix.recv(onelib.Message(msg), onelib.Sender(sender))
			onelib.Debug.Printf("%s: %s\n", displayName, msg.Text())
		} else {
			onelib.Debug.Println("Message: ", ev)
		}
		err = matrix.client.MarkRead(ev.RoomID, ev.ID)
		if err != nil {
			onelib.Error.Println(err)
		}
	})

	syncer.OnEventType("m.room.third_party_invite", func(ev *gomatrix.Event) {
		onelib.Debug.Println("Third Party Invite: ", ev)
	})
	syncer.OnEventType("m.room.member", func(ev *gomatrix.Event) {
		if ev.Content["membership"] != nil && ev.Content["membership"].(string) == "invite" {
			_, err = client.JoinRoom(ev.RoomID, "", nil)
			if err != nil {
				onelib.Error.Println(err)
			}
		} else {
			onelib.Debug.Println("Member: ", ev)
		}
	})

	client.SetDisplayName(onelib.DefaultNickname)
	// client.setAvatarToFile(onelib.DefaultAvatar) // TODO only do this if avatar hasn't been set yet

	err = matrix.client.SetStatus("online", "Test status.")
	/*_, err = matrix.client.SendStateEvent("<DM room ID>", "im.vector.user_status", matrixAuthUser, struct {
		Status string `json:"status"`
	}{"Test status"})*/
	if err != nil {
		onelib.Error.Println("Error setting presence:", err)
	}

	go matrix.handleconnections()
	return onelib.Protocol(matrix)
}

type matrixMessage struct {
	formattedText, text string
}

func (mm *matrixMessage) Text() string {
	return mm.text
}

func (mm *matrixMessage) FormattedText() string {
	return mm.formattedText
}

func (mm *matrixMessage) StripPrefix(prefix string) onelib.Message {
	if len(mm.text) > len(prefix) {
		prefix = prefix + " "
	}
	return onelib.Message(&matrixMessage{text: strings.Replace(mm.text, prefix, "", 1), formattedText: strings.Replace(mm.formattedText, prefix, "", 1)})
}

func (mm *matrixMessage) Raw() []byte {
	return []byte(mm.text)
}

type matrixSender struct {
	displayName, username string
	location              *matrixLocation
	uuid                  onelib.UUID
}

func (ms *matrixSender) DisplayName() string {
	return ms.displayName
}

func (ms *matrixSender) Username() string {
	return ms.username
}

func (ms *matrixSender) UUID() onelib.UUID {
	return ms.uuid
}

func (ms *matrixSender) Location() onelib.Location {
	return ms.location
}

func (ms *matrixSender) Protocol() string {
	return NAME
}

func (ms *matrixSender) Send(msg onelib.Message) {
	ms.location.Client.Send(ms.uuid, msg)
}

func (ms *matrixSender) SendText(text string) {
	ms.location.Client.SendText(ms.uuid, text)
}

func (ms *matrixSender) SendFormattedText(text, formattedText string) {
	ms.location.Client.SendFormattedText(ms.uuid, text, formattedText)
}

type matrixLocation struct {
	Client                                 *matrixClient // pointer to originating client
	displayName, nickname, topic, protocol string
	uuid                                   onelib.UUID
}

func (ml *matrixLocation) DisplayName() string {
	return ml.displayName
}

func (ml *matrixLocation) Nickname() string {
	return ml.nickname
}

func (ml *matrixLocation) Topic() string {
	return ml.topic
}

func (ml *matrixLocation) UUID() onelib.UUID {
	return ml.uuid
}

func (ml *matrixLocation) Send(msg onelib.Message) {
	ml.Client.Send(ml.uuid, msg)
}

func (ml *matrixLocation) SendText(text string) {
	ml.Client.SendText(ml.uuid, text)
}

func (ml *matrixLocation) SendFormattedText(text, formattedText string) {
	ml.Client.SendFormattedText(ml.uuid, text, formattedText)
}

func (ml *matrixLocation) Protocol() string {
	return NAME
}

type matrixClient struct {
	*gomatrix.Client
}

// Send sends a Message object to a location specified by to (usually a location or sender UUID).
func (mc *matrixClient) Send(to onelib.UUID, msg onelib.Message) {
	// code here
}

// SendText sends text to a location specified by to (usually a location or sender UUID).
func (mc *matrixClient) SendText(to onelib.UUID, text string) {
	_, err := mc.Client.SendText(string(to), text)
	if err != nil {
		onelib.Error.Println(err)
	}
}

// SendFormattedText sends formatted text to a location specified by to (usually a location or sender UUID).
func (mc *matrixClient) SendFormattedText(to onelib.UUID, text, formattedText string) {
	_, err := mc.SendMessageEvent(string(to), "m.room.message", &matrixProtocolMessage{Body: text, FormattedBody: formattedText, Format: "org.matrix.custom.html", Msgtype: "m.text"})
	if err != nil {
		onelib.Error.Println(err)
	}
}

// member contains useful data about a possible sender that's not typically sent in a message
type member struct {
	displayName string
}

type memberMap struct {
	mMap map[onelib.UUID]*member
	lock *sync.RWMutex
}

func (mm *memberMap) Get(uuid onelib.UUID) *member {
	mm.lock.RLock()
	mem := mm.mMap[uuid]
	mm.lock.RUnlock()
	return mem
}

func (mm *memberMap) Set(uuid onelib.UUID, mem *member) {
	mm.lock.Lock()
	mm.mMap[uuid] = mem
	mm.lock.Unlock()
}

// Matrix is the Protocol object used for handling anything Matrix related.
type Matrix struct {
	/*
	   Store useful data here such as connected rooms, admins, nickname, accepted prefixes, etc
	*/
	prefix       string
	nickname     string
	client       *matrixClient
	knownMembers *memberMap
}

// TODO finish this, only proof of concept right now
func (matrix *Matrix) handleconnections() {
	for {
		if err := matrix.client.Sync(); err != nil {
			onelib.Debug.Println("Sync() returned ", err)
		}
	}
}

// Name returns the name of the plugin, usually the filename.
func (matrix *Matrix) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (matrix *Matrix) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (matrix *Matrix) Version() string {
	return VERSION
}

// NewMessage should generate a message object from something
func (matrix *Matrix) NewMessage(raw []byte) onelib.Message {
	return nil
}

// Send sends a Message object to a location specified by to (usually a location or sender UUID).
func (matrix *Matrix) Send(to onelib.UUID, msg onelib.Message) {
	matrix.client.Send(to, msg)
}

// SendText sends text to a location specified by to (usually a location or sender UUID).
func (matrix *Matrix) SendText(to onelib.UUID, text string) {
	matrix.client.SendText(to, text)
}

// SendFormattedText sends formatted text to a location specified by to (usually a location or sender UUID).
func (matrix *Matrix) SendFormattedText(to onelib.UUID, text, formattedText string) {
	matrix.client.SendFormattedText(to, text, formattedText)
}

// GetUserDisplayName returns a user's display name from a UUID
//func (matrix *Matrix) GetUserDisplayName(uuid onelib.UUID) string

// recv should be called after you've recieved data and built a Message object
func (matrix *Matrix) recv(msg onelib.Message, sender onelib.Sender) {
	if string(sender.UUID()) != matrixAuthUser {
		onelib.ProcessMessage(matrix.prefix, msg, sender)
	}
}

// Remove currently doesn't do anything.
func (matrix *Matrix) Remove() {
	matrix.client.SetStatus("offline", "")
}
