// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/TheDiscordian/onebot/onelib"
)

const (
	// NAME is same as filename, minus extension
	NAME = "irc_bnetd"
	// LONGNAME is what's presented to the user
	LONGNAME = "IRC (bnetd)"
	// VERSION of the script
	VERSION = "v0.0.0"
)

var (
	// Nick for the bot to use (keep in mind bnetd requires this nick to correlate to an existing account)
	bnetNick string
	// Bnet server like "bnet.thedisco.zone:6667"
	bnetServer string
	// Password (likely required)
	bnetPass string
	// Channels to automatically join (comma separated)
	bnetAutoJoin string

	bnetConn net.Conn
)

func loadConfig() {
	bnetNick = onelib.GetTextConfig(NAME, "nick")
	bnetServer = onelib.GetTextConfig(NAME, "server")
	bnetPass = onelib.GetTextConfig(NAME, "pass")
	bnetAutoJoin = onelib.GetTextConfig(NAME, "auto_join")
}

// Load connects to BnetProtocol, and sets up listeners. It's required for OneBot.
func Load() onelib.Protocol {
	loadConfig()

	bnetClient := &BnetProtocol{prefix: onelib.DefaultPrefix}

	go handleConnection(bnetClient)

	return onelib.Protocol(bnetClient)
}

func handleConnection(c *BnetProtocol) {
	for {
		var err error
		bnetConn, err = net.Dial("tcp", bnetServer)
		if err != nil {
			onelib.Error.Println("[bnet]", err)
			time.Sleep(time.Second * 30)
			continue
		}
		bnetConn.Write([]byte(fmt.Sprintf("USER %s * * :%s\r\n", bnetNick, bnetNick)))
		bnetConn.Write([]byte(fmt.Sprintf("NICK %s\r\n", bnetNick)))
		r := bufio.NewReader(bnetConn)
		for err == nil {
			msgStr, err := r.ReadString('\n')
			if err != nil {
				break
			}
			msgStr = msgStr[:len(msgStr)-2]
			splitMsg := strings.Split(msgStr, " ")
			if len(splitMsg) < 2 {
				continue
			}
			if splitMsg[0] == "PING" {
				bnetConn.Write([]byte(fmt.Sprintf("PONG %s\r\n", splitMsg[1])))
			} else if splitMsg[1] == "001" {
				bnetConn.Write([]byte(fmt.Sprintf("PRIVMSG NICKSERV :identify %s\r\n", bnetPass)))
				if len(bnetAutoJoin) > 0 {
					joinSplit := strings.Split(bnetAutoJoin, ",")
					for _, r := range joinSplit {
						bnetConn.Write([]byte(fmt.Sprintf("JOIN %s\r\n", r)))
					}
				}
			} else if splitMsg[1] == "PRIVMSG" {
				onelib.Debug.Println("GOT MSG")
				splitMsg[3] = splitMsg[3][1:]
				onelib.Debug.Println(strings.Join(splitMsg[3:], " "))
				msg := &bnetMessage{text: strings.Join(splitMsg[3:], " ")}
				loc := &bnetLocation{displayName: splitMsg[2], uuid: onelib.UUID(splitMsg[2])}
				splitMsg[0] = splitMsg[0][1:]
				senderNick := strings.Split(splitMsg[0], "!")[0]
				sender := &bnetSender{displayName: senderNick, location: loc, uuid: onelib.UUID(splitMsg[0])}
				c.recv(msg, sender)
			}
			onelib.Debug.Println("[bnet]", msgStr)
		}
	}
}

// BnetProtocol is the Protocol object used for handling anything BnetProtocol related.
type BnetProtocol struct {
	/*
	   Store useful data here such as connected rooms, admins, nickname, accepted prefixes, etc
	*/
	prefix string
}

// Name returns the name of the plugin, usually the filename.
func (bp *BnetProtocol) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (bp *BnetProtocol) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (bp *BnetProtocol) Version() string {
	return VERSION
}

// NewMessage should generate a message object from something
func (bp *BnetProtocol) NewMessage(raw []byte) onelib.Message {
	return nil
}

// Send sends a Message object to a location specified by to (usually a location or sender UUID).
func (bp *BnetProtocol) Send(to onelib.UUID, msg onelib.Message) {
	bnetSendText(to, msg.Text())
}

// SendText sends text to a location specified by to (usually a location or sender UUID).
func (bp *BnetProtocol) SendText(to onelib.UUID, text string) {
	bnetSendText(to, text)
}

// SendFormattedText sends formatted text to a location specified by to (usually a location or sender UUID).
func (bp *BnetProtocol) SendFormattedText(to onelib.UUID, text, formattedText string) {
	bnetSendText(to, text)
}

// recv should be called after you've recieved data and built a Message object
func (bp *BnetProtocol) recv(msg onelib.Message, sender onelib.Sender) {
	onelib.ProcessMessage(bp.prefix, msg, sender)
}

// Remove should disconnect any open connections making it so the bot can forget about the protocol cleanly.
func (bp *BnetProtocol) Remove() {
	/*
	   Unload code goes here (disconnects)
	*/
}

type bnetMessage struct {
	text string
}

func (bm *bnetMessage) Text() string {
	return bm.text
}

func (bm *bnetMessage) FormattedText() string {
	return bm.text
}

func (bm *bnetMessage) StripPrefix(prefix string) onelib.Message {
	if len(bm.text) > len(prefix) {
		prefix = prefix + " "
	}
	return onelib.Message(&bnetMessage{text: strings.Replace(bm.text, prefix, "", 1)})
}

func (bm *bnetMessage) Raw() []byte {
	return []byte(bm.text)
}

func bnetSendText(to onelib.UUID, text string) {
	lines := strings.Split(text, "\n")
	for _, msg := range lines {
		_, err := bnetConn.Write([]byte(fmt.Sprintf("PRIVMSG %s :%s\r\n", string(to), msg)))
		if err != nil {
			onelib.Error.Println(err)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

type bnetSender struct {
	displayName string
	location    *bnetLocation
	uuid        onelib.UUID
}

func (bs *bnetSender) DisplayName() string {
	return bs.displayName
}

func (bs *bnetSender) Username() string {
	return bs.displayName
}

func (bs *bnetSender) UUID() onelib.UUID {
	return bs.uuid
}

func (bs *bnetSender) Location() onelib.Location {
	return bs.location
}

func (bs *bnetSender) Protocol() string {
	return NAME
}

func (bs *bnetSender) Send(msg onelib.Message) {
	bnetSendText(bs.uuid, msg.Text())
}

func (bs *bnetSender) SendText(text string) {
	bnetSendText(bs.uuid, text)
}

func (bs *bnetSender) SendFormattedText(text, formattedText string) {
	bnetSendText(bs.uuid, text)
}

type bnetLocation struct {
	displayName string
	uuid        onelib.UUID
}

func (bl *bnetLocation) DisplayName() string {
	return bl.displayName
}

func (bl *bnetLocation) Nickname() string {
	return bl.displayName
}

func (bl *bnetLocation) Topic() string {
	return ""
}

func (bl *bnetLocation) UUID() onelib.UUID {
	return bl.uuid
}

func (bl *bnetLocation) Send(msg onelib.Message) {
	bnetSendText(bl.uuid, msg.Text())
}

func (bl *bnetLocation) SendText(text string) {
	bnetSendText(bl.uuid, text)
}

func (bl *bnetLocation) SendFormattedText(text, formattedText string) {
	bnetSendText(bl.uuid, text)
}

func (bl *bnetLocation) Protocol() string {
	return NAME
}
