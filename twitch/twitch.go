package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// https://github.com/videolan/vlc/blob/master/share/lua/playlist/twitch.lua#L81

// Client manages communication with the twitch API.
type Client struct {
	client      *http.Client
	clientID    string
	apiURL      string
	usherAPIURL string
}

// New returns a new twitch API client.
func New(client *http.Client, clientID string) Client {
	return Client{client, clientID, "https://api.twitch.tv/", "http://usher.twitch.tv/"}
}

// Custom returns a new twitch API client with custom API endpoints
func Custom(client *http.Client, clientID, apiURL, usherAPIURL string) Client {
	return Client{client, clientID, apiURL, usherAPIURL}
}

type token struct {
	Token string `json:"token"`
	Sig   string `json:"sig"`
}

func (c *Client) vodToken(ctx context.Context, id string) (token, error) {
	u := fmt.Sprintf("%sapi/vods/%s/access_token?client_id=%s", c.apiURL, id, c.clientID)
	resp, err := c.client.Get(u)
	if err != nil {
		return token{}, errors.WithStack(err)
	}
	defer resp.Body.Close()
	if s := resp.StatusCode; s < 200 || s >= 300 {
		b, _ := ioutil.ReadAll(resp.Body)
		return token{}, errors.Errorf("%d\n%s\n%s", resp.StatusCode, u, string(b))
	}
	var t token
	return t, errors.WithStack(json.NewDecoder(resp.Body).Decode(&t))
}

// VODM3U8 retrieves the M3U8 file of a specific VOD.
func (c *Client) VODM3U8(ctx context.Context, id string) ([]byte, error) {
	tok, err := c.vodToken(ctx, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	u := fmt.Sprintf("%svod/%s?nauth=%s&nauthsig=%s&allow_audio_only=true&allow_source=true",
		c.usherAPIURL, id, tok.Token, tok.Sig)
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()
	if s := resp.StatusCode; s < 200 || s >= 300 {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%d\n%s\n%s", resp.StatusCode, u, string(b))
	}
	return ioutil.ReadAll(resp.Body)
}
