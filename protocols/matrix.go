// Copyright (c) 2020, The OneBot Contributors. All rights reserved.
package main

import (
	"strings"

	"github.com/TheDiscordian/onebot/onelib"
	"github.com/matrix-org/gomatrix"
)

const (
	// Same as filename, minus extension
	NAME = "matrix"
	// Name presented to user
	LONGNAME = "Matrix"
	// Version of the script (higher regarded as newer)
	VERSION = 0
	// TODO move all these constants to the config...
	// MATRIX_HOME_SERVER
	MATRIX_HOME_SERVER = ""
	// MATRIX_AUTH_USER
	MATRIX_AUTH_USER = ""
	// MATRIX_AUTH_TOKEN if blank, falls back onto pass
	MATRIX_AUTH_TOKEN = ""
	// MATRIX_AUTH_PASS
	MATRIX_AUTH_PASS = ""
)

// TODO store rooms as a map of locations & a map of senders, mapped by UID
func Load() onelib.Protocol {
	/*
	   Code to be executed on-load goes here (connects)
	*/
	client, err := gomatrix.NewClient(MATRIX_HOME_SERVER, MATRIX_AUTH_USER, MATRIX_AUTH_TOKEN)
	if err != nil {
		onelib.Error.Panicln(err)
	}
	if MATRIX_AUTH_TOKEN == "" {
		if MATRIX_AUTH_USER == "" {
			panic("both MATRIX_AUTH_USER and MATRIX_AUTH_TOKEN can't be blank.")
		}
		resp, err := client.Login(&gomatrix.ReqLogin{
			Type:     "m.login.password",
			User:     MATRIX_AUTH_USER,
			Password: MATRIX_AUTH_PASS,
		})
		if err != nil {
			onelib.Error.Panicln(err)
		}
		// TODO automatically save access token
		onelib.Info.Println("Access token:", resp.AccessToken)
		client.SetCredentials(resp.UserID, resp.AccessToken)
	}
	syncer := client.Syncer.(*gomatrix.DefaultSyncer)

	matrix := &Matrix{client: &matrixClient{client: client}, prefix: onelib.DefaultPrefix, nickname: onelib.DefaultNickname}

	syncer.OnEventType("m.room.message", func(ev *gomatrix.Event) {
		if ev.Content["msgtype"] != nil && ev.Content["msgtype"].(string) == "m.text" {
			msg := &matrixMessage{text: ev.Content["body"].(string)}
			mc := &matrixClient{client: client}
			ml := &matrixLocation{Client: mc, uuid: onelib.UUID(ev.RoomID)}
			sender := &matrixSender{uuid: onelib.UUID(ev.Sender), location: ml}
			matrix.recv(onelib.Message(msg), onelib.Sender(sender))
			onelib.Debug.Printf("%s: %s\n", ev.Sender, msg.Text())
		} else {
			onelib.Debug.Println("Message: ", ev.Sender)
		}
		urlPath := client.BuildURL("rooms", ev.RoomID, "receipt", "m.read", ev.ID)
		err = client.MakeRequest("POST", urlPath, nil, nil)
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

	go matrix.handleconnections()
	return onelib.Protocol(matrix)
}

type matrixMessage struct {
	text string
}

func (mm *matrixMessage) Text() string {
	return mm.text
}

func (mm *matrixMessage) StripPrefix() onelib.Message {
	return onelib.Message(&matrixMessage{text: strings.Join(strings.Split(mm.text, " ")[1:], " ")})
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
	onelib.Error.Println("not implemented.")
}

func (ms *matrixSender) SendText(text string) {
	ms.location.Client.SendText(ms.uuid, text)
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
	onelib.Error.Println("not implemented.")
}

func (ml *matrixLocation) SendText(text string) {
	ml.Client.SendText(ml.uuid, text)
}

func (ml *matrixLocation) Protocol() string {
	return NAME
}

type matrixClient struct {
	client *gomatrix.Client
}

// Send sends a Message object to a location specified by to (usually a location or sender UUID).
func (mc *matrixClient) Send(to onelib.UUID, msg onelib.Message) {
	// code here
}

// SendText sends text to a location specified by to (usually a location or sender UUID).
func (mc *matrixClient) SendText(to onelib.UUID, text string) {
	_, err := mc.client.SendText(string(to), text)
	if err != nil {
		onelib.Error.Println(err)
	}
}

func (mc *matrixClient) Sync() error {
	return mc.client.Sync()
}

type Matrix struct {
	/*
	   Store useful data here such as connected rooms, admins, nickname, accepted prefixes, etc
	*/
	prefix   string
	nickname string
	client   *matrixClient
}

// TODO finish this, only proof of concept right now
func (matrix *Matrix) handleconnections() {
	for {
		if err := matrix.client.Sync(); err != nil {
			onelib.Debug.Println("Sync() returned ", err)
		}
	}
}

func (matrix *Matrix) Name() string {
	return NAME
}

func (matrix *Matrix) LongName() string {
	return LONGNAME
}

func (matrix *Matrix) Version() int {
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

// recv should be called after you've recieved data and built a Message object
func (matrix *Matrix) recv(msg onelib.Message, sender onelib.Sender) {
	onelib.Debug.Println("Got msg:", msg, "from:", sender)
	if string(sender.UUID()) != MATRIX_AUTH_USER {
		onelib.ProcessMessage(matrix.prefix, msg, sender)
	}
}

func (matrix *Matrix) Remove() {
	/*
	   Unload code goes here (disconnects)
	*/
}
