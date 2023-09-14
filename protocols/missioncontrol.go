// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"crypto/sha256"
	"net/http"
	"html/template"
	"fmt"
	"io/ioutil"
	"crypto/rand"
	"encoding/hex"

	"github.com/TheDiscordian/onebot/onelib"
)

const (
	// NAME is same as filename, minus extension
	NAME = "missioncontrol"
	// LONGNAME is what's presented to the user
	LONGNAME = "Mission Control"
	// VERSION of the script
	VERSION = "v0.0.0"
)

var (
	MissionControlPort int
	Users *users // FIXME: This isn't thread-safe, make it thread-safe
)

type users struct {
	Users []*user
}

type user struct {
	Username string // Plaintext username
	Password [32]byte // Hashed password
	Session string // Session token (TODO make this expire)
}

func loadConfig() {
	// MissionControlServer = onelib.GetTextConfig(NAME, "server")
	MissionControlPort, _ = onelib.GetIntConfig(NAME, "port")
	onelib.Db.GetObj(NAME, "users", &Users)
}

// Load connects to MissionControl, and sets up listeners. It's required for OneBot.
func Load() onelib.Protocol {
	loadConfig()
	/*
	   Code to be executed on-load goes here (connects)
	*/
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "protocols/missioncontrol/style.css")
	})

	go http.ListenAndServe(fmt.Sprintf("localhost:%d", MissionControlPort), nil)
	return onelib.Protocol(&MissionControl{})
}

func GenerateSecureToken(length int) string {
    b := make([]byte, length)
    if _, err := rand.Read(b); err != nil {
        return ""
    }
    return hex.EncodeToString(b)
}

func servePage(w http.ResponseWriter, r *http.Request, page string) {
	// Load template from protocols/missioncontrol/page.tmpl
	index, err := ioutil.ReadFile("protocols/missioncontrol/"+page+".tmpl")
	if err != nil {
		onelib.Error.Println(err)
		fmt.Fprintf(w, "Internal server error.")
		return
	}
	indexTpl, err := template.New("index").Parse(string(index))
	if err != nil {
		onelib.Error.Println(err)
		fmt.Fprintf(w, "Internal server error.")
		return
	}

	indexTpl.ParseFiles("protocols/missioncontrol/header.tmpl", "protocols/missioncontrol/footer.tmpl")

	indexVars := struct {
		PluginCount int
		ProtocolCount int
		Version string
	}{
		PluginCount: len(onelib.Plugins.List()),
		ProtocolCount: len(onelib.Protocols.List()),
		Version: onelib.VERSION,
	}

	err = indexTpl.Execute(w, indexVars)
	if err != nil {
		onelib.Error.Println(err)
		fmt.Fprintf(w, "Internal server error.")
		return
	}
}

func userMatchesSession(username, session string) bool {
	for _, user := range Users.Users {
		if user.Username == username && user.Session == session {
			return true
		}
	}
	return false
}

func userMatchesPassword(username, password string) bool {
	for _, user := range Users.Users {
		if user.Username == username && user.Password == sha256.Sum256([]byte(password)) {
			return true
		}
	}
	return false
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	// Check if the cookie "session" exists and contains the value "pickles"
	ses, err := r.Cookie("session")
	if err != nil || Users == nil || len(Users.Users) == 0 {
		serveLogin(w, r)
		return
	}
	user, err := r.Cookie("username")
	if err != nil || !userMatchesSession(user.Value, ses.Value) {
		serveLogin(w, r)
		return
	}
	servePage(w, r, "index")
}

func serveLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password") // z1I1^ucnOiuaYG&hvbbFe2OJ
	if Users == nil || len(Users.Users) == 0 {
		if len(username) > 0 && len(password) > 11 {
			Users = &users{
				Users: []*user{
					&user{
						Username: username,
						Password: sha256.Sum256([]byte(password)),
						Session: GenerateSecureToken(32),
					},
				},
			}
			onelib.Db.PutObj(NAME, "users", Users)
			http.SetCookie(w, &http.Cookie{Name: "session", Value: Users.Users[0].Session})
			http.SetCookie(w, &http.Cookie{Name: "username", Value: Users.Users[0].Username})
			servePage(w, r, "index")
		}
		serveFirstLogin(w, r)
		return
	}

	if userMatchesPassword(username, password) {
		// Generate a session token
		session := GenerateSecureToken(32)
		// Set a cookie with the session token
		http.SetCookie(w, &http.Cookie{Name: "session", Value: session})
		http.SetCookie(w, &http.Cookie{Name: "username", Value: username})
		// Set the session token in the Users object
		for _, user := range Users.Users {
			if user.Username == username {
				user.Session = session
				break
			}
		}
		// Save the Users object to the database
		onelib.Db.PutObj(NAME, "users", Users)
		servePage(w, r, "index")
		return
	}
	servePage(w, r, "login")
}

func serveFirstLogin(w http.ResponseWriter, r *http.Request) {
	servePage(w, r, "firstlogin")
}

// MissionControl is the Protocol object used for handling anything MissionControl related.
type MissionControl struct {
	/*
	   Store useful data here such as connected rooms, admins, nickname, accepted prefixes, etc
	*/
}

// Name returns the name of the plugin, usually the filename.
func (mc *MissionControl) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (mc *MissionControl) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (mc *MissionControl) Version() string {
	return VERSION
}

// NewMessage should generate a message object from something
func (mc *MissionControl) NewMessage(raw []byte) onelib.Message {
	return nil
}

// Send sends a Message object to a location specified by to (usually a location or sender UUID).
func (mc *MissionControl) Send(to onelib.UUID, msg onelib.Message) {
	// code here
}

// SendText sends text to a location specified by to (usually a location or sender UUID).
func (mc *MissionControl) SendText(to onelib.UUID, text string) {
	// code here
}

// SendFormattedText sends formatted text to a location specified by to (usually a location or sender UUID).
func (mc *MissionControl) SendFormattedText(to onelib.UUID, text, formattedText string) {
	// code here
}

// recv should be called after you've recieved data and built a Message object
func (mc *MissionControl) recv(msg onelib.Message, sender onelib.Sender) {
	onelib.ProcessMessage([]string{onelib.DefaultPrefix}, msg, sender)
}

// Remove should disconnect any open connections making it so the bot can forget about the protocol cleanly.
func (mc *MissionControl) Remove() {
	/*
	   Unload code goes here (disconnects)
	*/
}