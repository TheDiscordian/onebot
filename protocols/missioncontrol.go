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
	"sync"

	"github.com/TheDiscordian/onebot/libs/missioncontrol"
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
	users map[string]*user // Public to make setting easier, but this should only be accessed directly for creation
	lock *sync.Mutex
}

func (u *users) Set(username string, user *user) {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.users[username] = user
	onelib.Db.PutObj(NAME, "users", u.users)
}

func (u *users) Get(username string) *user {
	u.lock.Lock()
	defer u.lock.Unlock()
	return u.users[username]
}

func (u *users) Del(username string) {
	u.lock.Lock()
	defer u.lock.Unlock()
	delete(u.users, username)
	onelib.Db.PutObj(NAME, "users", u.users)
}

func (u *users) List() []string {
	u.lock.Lock()
	defer u.lock.Unlock()
	users := make([]string, 0, len(u.users))
	for username := range u.users {
		users = append(users, username)
	}
	return users
}

type user struct {
	Password [32]byte // Hashed password
	Session string // Session token (TODO make this expire)
}

func loadConfig() {
	// MissionControlServer = onelib.GetTextConfig(NAME, "server")
	MissionControlPort, _ = onelib.GetIntConfig(NAME, "port")
	Users = new(users)
	onelib.Db.GetObj(NAME, "users", &Users.users)
	Users.lock = new(sync.Mutex)
	if Users.users == nil {
		Users.users = make(map[string]*user)
	}
	missioncontrol.Init()
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
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/login", serveLogin)
	http.HandleFunc("/settings", serveSettings)
	http.HandleFunc("/plugins", servePlugins)

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

func servePage(w http.ResponseWriter, r *http.Request, page string, loggedIn bool) {
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

	var (
		pluginCount, protocolCount int
		plugins []string
	)
	if loggedIn {
		switch page {
		case "index":
			pluginCount = len(onelib.Plugins.List())
			protocolCount = len(onelib.Protocols.List())
		case "plugins":
			plugins = missioncontrol.Plugins.List()
		}
	}

	indexVars := struct {
		PluginCount int   // Count of all plugins loaded by OneBot
		ProtocolCount int // Count of all protocol plugins loaded by OneBot
		Version string
		LoggedIn bool
		Users []string    // List of users registered with Mission Control
		Plugins []string  // List of plugins loaded which support Mission Control
	}{
		PluginCount: pluginCount,
		ProtocolCount: protocolCount,
		Version: onelib.VERSION,
		LoggedIn: loggedIn,
		Users: Users.List(),
		Plugins: plugins,
	}

	err = indexTpl.Execute(w, indexVars)
	if err != nil {
		onelib.Error.Println(err)
		fmt.Fprintf(w, "Internal server error.")
		return
	}
}

func userMatchesSession(username, session string) bool {
	user := Users.Get(username)
	if user != nil && user.Session == session {
		return true
	}
	return false
}

func userMatchesPassword(username, password string) bool {
	user := Users.Get(username)
	if user != nil && user.Password == sha256.Sum256([]byte(password)) {
		return true
	}
	return false
}

func loggedIn(r *http.Request) bool {
	ses, err := r.Cookie("session")
	if err != nil || len(Users.List()) == 0 {
		return false
	}
	userCookie, err := r.Cookie("username")
	if err != nil || !userMatchesSession(userCookie.Value, ses.Value) {
		return false
	}
	return true
}

func serveSettings(w http.ResponseWriter, r *http.Request) {
	if !loggedIn(r) {
		serveLogin(w, r)
		return
	}
	servePage(w, r, "settings", true)
}

func servePlugins(w http.ResponseWriter, r *http.Request) {
	if !loggedIn(r) {
		serveLogin(w, r)
		return
	}
	servePage(w, r, "plugins", true)
}

func logout(w http.ResponseWriter, r *http.Request) {
	ses, err := r.Cookie("session")
	if err != nil || len(Users.List()) == 0 {
		serveLogin(w, r)
		return
	}
	userCookie, err := r.Cookie("username")
	if err != nil || !userMatchesSession(userCookie.Value, ses.Value) {
		serveLogin(w, r)
		return
	}
	u := Users.Get(userCookie.Value)
	if u == nil {
		serveLogin(w, r)
		return
	}
	u.Session = ""
	Users.Set(userCookie.Value, u)
	// Expire the cookie on the user's end too
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", MaxAge: -1})
	serveLogin(w, r)
	return
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	ses, err := r.Cookie("session")
	if err != nil || len(Users.List()) == 0 {
		serveLogin(w, r)
		return
	}
	user, err := r.Cookie("username")
	if err != nil || !userMatchesSession(user.Value, ses.Value) {
		serveLogin(w, r)
		return
	}
	servePage(w, r, "index", true)
}

func serveLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	if len(Users.List()) == 0 {
		if len(username) > 0 && len(password) > 11 {
			user := &user{
						Password: sha256.Sum256([]byte(password)),
						Session: GenerateSecureToken(32),
					}
			Users.Set(username, user)
			http.SetCookie(w, &http.Cookie{Name: "session", Value: user.Session, SameSite: http.SameSiteStrictMode, Secure: true, HttpOnly: true})
			http.SetCookie(w, &http.Cookie{Name: "username", Value: username, SameSite: http.SameSiteStrictMode})
			servePage(w, r, "index", true)
			return
		}
		serveFirstLogin(w, r)
		return
	}

	if userMatchesPassword(username, password) {
		// Generate a session token
		session := GenerateSecureToken(32)
		// Set a cookie with the session token
		http.SetCookie(w, &http.Cookie{Name: "session", Value: session, SameSite: http.SameSiteStrictMode, Secure: true, HttpOnly: true})
		http.SetCookie(w, &http.Cookie{Name: "username", Value: username, SameSite: http.SameSiteStrictMode})
		// Set the session token in the Users object
		u := Users.Get(username)
		if u == nil {
			fmt.Fprintf(w, "Internal server error.")
			return
		}
		u.Session = session
		Users.Set(username, u)
		servePage(w, r, "index", true)
		return
	}
	servePage(w, r, "login", false)
}

func serveFirstLogin(w http.ResponseWriter, r *http.Request) {
	servePage(w, r, "firstlogin", false)
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