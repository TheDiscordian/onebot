// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/TheDiscordian/onebot/onelib"
)

const (
	// NAME is same as filename, minus extension
	NAME = "ipfs"
	// LONGNAME is what's presented to the user
	LONGNAME = "IPFS Plugin"
	// VERSION of the plugin
	VERSION = "v0.0.0"
)

// Load returns the Plugin object.
func Load() onelib.Plugin {
	return new(IPFSPlugin)
}

type IPFSCheckResp struct {
	ConnectionError          string
	PeerFoundInDHT           map[string]int
	CidInDHT                 bool
	DataAvailableOverBitswap IPFSBitswapResp
}

type IPFSBitswapResp struct {
	Error     string
	Responded bool
	Found     bool
}

type IPFSFindProvsResp struct {
	ID        string
	Responses []*IPFSPeer
	Type      int
}

type IPFSPeer struct {
	Addrs []string
	ID    string
}

func doRequest(timeout time.Duration, url string) ([]byte, error) {
	var cancel context.CancelFunc
	ctx := context.Background()
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	c := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func doIPFSCheckRequest(timeout time.Duration, url string) (*IPFSCheckResp, error) {
	body, err := doRequest(timeout, url)
	if err != nil {
		return nil, err
	}

	out := new(IPFSCheckResp)
	err = json.Unmarshal(body, out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func doIPFSFindProvsRequest(timeout time.Duration, url string) (int, error) {
	var cancel context.CancelFunc
	ctx := context.Background()
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	c := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	if err != nil {
		return 0, err
	}

	providers := 0
	peers := make(map[string]int, 20)

	// Decode the json stream and process it
	for dec.More() {
		out := new(IPFSFindProvsResp)
		err := dec.Decode(out)
		if err != nil {
			return providers, err
		}
		if out.Type == 4 { // type 4 is "found" I guess
			for _, peer := range out.Responses {
				peers[peer.ID] = 0
			}
		}
	}

	return len(peers), nil
}

func ipfsDHTFindProvs(msg onelib.Message, sender onelib.Sender) {
	const USAGE = "Usage: ipfs-findprovs <CID>"
	txt := msg.Text()
	if len(txt) <= 1 {
		sender.Location().SendText(USAGE)
		return
	}

	sender.Location().SendText("Checking DHT for " + txt + " (up to 30s)")
	providers, err := doIPFSFindProvsRequest(time.Second*30, "http://127.0.0.1:5001/api/v0/dht/findprovs?arg="+txt)
	if err != nil && providers == 0 {
		onelib.Debug.Println(err)
		sender.Location().SendText(fmt.Sprintf("No providers found for %s within 30s.", txt))
		return
	}

	sender.Location().SendText(fmt.Sprintf("%d providers found for %s.", providers, txt))
}

func ipfsCheck(msg onelib.Message, sender onelib.Sender) {
	const USAGE = "Usage: ipfs-check <multiaddr> <CID> [Backend URL]"
	splitTxt := strings.Split(msg.Text(), " ")
	if len(splitTxt) < 2 || len(splitTxt) > 3 {
		sender.Location().SendText(USAGE)
		return
	}
	multiaddr := splitTxt[0]
	cid := splitTxt[1]
	backend := "https://ipfs-check-backend.ipfs.io"
	if len(splitTxt) == 3 {
		backend = splitTxt[2]
	}
	peerIdIndex := strings.LastIndex(multiaddr, "/p2p/")
	if peerIdIndex < 0 || len(multiaddr)-10 < peerIdIndex {
		sender.Location().SendText("Multiaddr appears malformed (can't find peerId)")
		sender.Location().SendText(USAGE)
		return
	}
	// peerId := multiaddr[peerIdIndex+5:]
	addrPart := multiaddr[:peerIdIndex]
	out, err := doIPFSCheckRequest(time.Minute, backend+"?multiaddr="+multiaddr+"&cid="+cid)
	if err != nil {
		onelib.Error.Println("[IPFS] " + err.Error())
		sender.Location().SendText("Error parsing response: " + err.Error())
		return
	}

	var resp string

	if out.ConnectionError != "" {
		resp += "❌ Could not connect to multiaddr: " + out.ConnectionError + "\n"
	} else {
		resp += "✅ Successfully connected to multiaddr\n"
	}

	var foundAddr bool
	for key, _ := range out.PeerFoundInDHT {
		if key == addrPart {
			foundAddr = true
			resp += "✅ Found multiaddr with " + strconv.Itoa(out.PeerFoundInDHT[key]) + " dht peers\n"
			break
		}
	}
	if !foundAddr {
		resp += "❌ Could not find the given multiaddr in the dht.\n" // TODO consider adding in "Instead found: [...]"
	}

	if out.CidInDHT {
		resp += "✅ Found multihash adverised in the dht\n"
	} else {
		resp += "❌ The peer responded that it does not have the CID\n"
	}

	if out.DataAvailableOverBitswap.Error != "" {
		resp += "❌ There was an error downloading the CID from the peer: " + out.DataAvailableOverBitswap.Error + "\n"
	} else {
		if !out.DataAvailableOverBitswap.Responded {
			resp += "❌ The peer did not quickly respond if it had the CID\n"
		} else {
			if out.DataAvailableOverBitswap.Found {
				resp += "✅ The peer responded that it has the CID\n"
			} else {
				resp += "❌ The peer responded that it does not have the CID\n"
			}
		}
	}

	sender.Location().SendText(resp)
}

// IPFSPlugin is an object for satisfying the Plugin interface.
type IPFSPlugin int

// Name returns the name of the plugin, usually the filename.
func (ip *IPFSPlugin) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (ip *IPFSPlugin) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (ip *IPFSPlugin) Version() string {
	return VERSION
}

// Implements returns a map of commands and monitor the plugin implements.
func (ip *IPFSPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"ipfs-check": ipfsCheck, "ipfs-findprovs": ipfsDHTFindProvs}, nil
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (ip *IPFSPlugin) Remove() {
}
