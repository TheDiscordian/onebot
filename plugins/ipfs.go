// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

var (
	sharedWeb3StorageKey string
)

func loadConfig() {
	sharedWeb3StorageKey = onelib.GetTextConfig(NAME, "shared_web3_storage_key")
}

// Load returns the Plugin object.
func Load() onelib.Plugin {
	loadConfig()
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

func doRequest(timeout time.Duration, url string, limit int64) ([]byte, error) {
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

	reader := resp.Body
	if limit >= 0 {
		reader = io.NopCloser(&io.LimitedReader{R: resp.Body, N: limit})
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func doWeb3Request(timeout time.Duration, url string, data []byte) ([]byte, error) {
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
	req.Header.Set("Authorization", "Bearer "+sharedWeb3StorageKey)
	req.Body = io.NopCloser(bytes.NewReader(data))
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func doIPFSCheckRequest(timeout time.Duration, url string) (*IPFSCheckResp, error) {
	body, err := doRequest(timeout, url, -1)
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
		sender.Location().SendText(fmt.Sprintf("No providers found for %s within 30s.", txt))
		return
	}

	sender.Location().SendText(fmt.Sprintf("%d providers found for %s.", providers, txt))
}

func ipfsBlockStat(msg onelib.Message, sender onelib.Sender) {
	const USAGE = "Usage: ipfs-stat <CID>"
	txt := msg.Text()
	if len(txt) <= 1 {
		sender.Location().SendText(USAGE)
		return
	}

	sender.Location().SendText("Trying to stat " + txt + " (up to 30s)")
	body, err := doRequest(time.Second*30, "http://127.0.0.1:5001/api/v0/block/stat?arg="+txt, -1)
	if err != nil || string(body) == "" {
		sender.Location().SendText(fmt.Sprintf("Failed to retrieve %s within 30s.", txt))
		return
	}

	sender.Location().SendText(fmt.Sprintf("Successfully retrieved %s.", txt))
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
		resp += "❌ Could not find the multihash in the dht\n"
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

func web3Help(msg onelib.Message, sender onelib.Sender) {
	sender.Location().SendText(`Commands:
* web3-store <cid> (Store up to 100MB on web3.storage)

More features coming soon like being able to set your own API key and bypass the 100MB limit, stay tuned 🚀`)
}

func web3Store(msg onelib.Message, sender onelib.Sender) {
	const USAGE = "Usage: web3-store <cid> (100MB limit)"
	if sharedWeb3StorageKey == "" {
		sender.Location().SendText("This feature hasn't been enabled by the owner of this bot, please contact them for details.")
		return
	}
	if msg.Text() == "" {
		sender.Location().SendText(USAGE)
		return
	}
	sender.Location().SendText("Beginning download of '" + msg.Text() + "' 🚀")
	data, err := doRequest(time.Second*120, "http://127.0.0.1:5001/api/v0/dag/export?arg="+msg.Text(), 100000000) // 100MB limit
	if err != nil {
		sender.Location().SendText(fmt.Sprintf("Error exporting CID: %s\n\n%s", err.Error(), USAGE))
		return
	}
	sender.Location().SendText(fmt.Sprintf("Got %dMiB of data! Uploading to web3.storage...", len(data)/1048576))
	resp, err := doWeb3Request(time.Second*120, "https://api.web3.storage/car", data)
	if err != nil {
		sender.Location().SendText(fmt.Sprintf("Error uploading CID: %s\n\n%s", err.Error(), USAGE))
		return
	}
	sender.Location().SendText("Upload complete! web3.storage responded: " + string(resp))
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
	return map[string]onelib.Command{"ipfs-check": ipfsCheck, "ipfs-findprovs": ipfsDHTFindProvs, "ipfs-stat": ipfsBlockStat, "web3-help": web3Help, "web3-store": web3Store}, nil
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (ip *IPFSPlugin) Remove() {
}
