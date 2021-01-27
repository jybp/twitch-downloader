package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"github.com/pkg/errors"
)

// ID extract the ID from a VOD url.
func ID(URL string) (string, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return "", errors.WithStack(err)
	}
	if u.Hostname() != "www.twitch.tv" {
		return "", errors.New("URL host is not twitch.tv:" + u.Hostname())
	}
	if !strings.HasPrefix(u.Path, "/videos/") {
		return "", errors.New("URL path does not contain /videos/")
	}
	_, id := path.Split(u.Path)
	return id, nil
}

// Client manages communication with the twitch API.
type Client struct {
	client      *http.Client
	clientID    string
	apiURL      string
	usherAPIURL string
}

// New returns a new twitch API client.
func New(client *http.Client, clientID string) Client {
	return Client{client, clientID, "https://gql.twitch.tv/gql", "http://usher.twitch.tv/"}
}

// Custom returns a new twitch API client with custom API endpoints
func Custom(client *http.Client, clientID, apiURL, usherAPIURL string) Client {
	return Client{client, clientID, apiURL, usherAPIURL}
}

func (c *Client) vodToken(ctx context.Context, id string) (token, sig string, err error) {
	gqlPayload := `{"operationName":"PlaybackAccessToken_Template","query":"query PlaybackAccessToken_Template($login: String!, $isLive: Boolean!, $vodID: ID!, $isVod: Boolean!, $playerType: String!) {  streamPlaybackAccessToken(channelName: $login, params: {platform: \"web\", playerBackend: \"mediaplayer\", playerType: $playerType}) @include(if: $isLive) {    value    signature    __typename  }  videoPlaybackAccessToken(id: $vodID, params: {platform: \"web\", playerBackend: \"mediaplayer\", playerType: $playerType}) @include(if: $isVod) {    value    signature    __typename  }}","variables":{"isLive":false,"login":"","isVod":true,"vodID":"%s","playerType":"site"}}`
	body := strings.NewReader(fmt.Sprintf(gqlPayload, id))
	req, err := http.NewRequest(http.MethodPost, c.apiURL, body)
	if err != nil {
		return "", "", errors.WithStack(err)
	}
	req.Header.Set("Client-Id", c.clientID)
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return "", "", errors.WithStack(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", errors.Errorf("%v\n%s", err, string(dump))
	}
	defer resp.Body.Close()
	if s := resp.StatusCode; s < 200 || s >= 300 {
		return "", "", errors.Errorf("invalid status code %d\n%s", s, string(dump))
	}

	type respPayload struct {
		Data struct {
			VideoPlaybackAccessToken struct {
				Value     string `json:"value"`
				Signature string `json:"signature"`
			} `json:"videoPlaybackAccessToken"`
		} `json:"data"`
	}

	var p respPayload
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return "", "", errors.Errorf("%v\n%s", err, string(dump))
	}
	return p.Data.VideoPlaybackAccessToken.Value, p.Data.VideoPlaybackAccessToken.Signature, nil
}

// M3U8 retrieves the M3U8 file of a specific VOD.
func (c *Client) M3U8(ctx context.Context, id string) ([]byte, error) {
	tok, sig, err := c.vodToken(ctx, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	u := fmt.Sprintf("%svod/%s?nauth=%s&nauthsig=%s&allow_audio_only=true&allow_source=true",
		c.usherAPIURL, id, tok, sig)
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()
	if s := resp.StatusCode; s < 200 || s >= 300 {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.Errorf("%d\n%s\n%s", s, u, string(b))
	}
	return ioutil.ReadAll(resp.Body)
}

// VOD describes a twitch VOD.
type VOD struct {
	Title string
}

// VOD retrieves the video informations of a specific VOD.
func (c *Client) VOD(ctx context.Context, id string) (VOD, error) {
	gqlPayload := `{"operationName":"VideoMetadata","variables":{"channelLogin":"","videoID":"%s"},"extensions":{"persistedQuery":{"version":1,"sha256Hash":"226edb3e692509f727fd56821f5653c05740242c82b0388883e0c0e75dcbf687"}}}`
	body := strings.NewReader(fmt.Sprintf(gqlPayload, id))
	req, err := http.NewRequest(http.MethodPost, c.apiURL, body)
	if err != nil {
		return VOD{}, errors.WithStack(err)
	}
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return VOD{}, errors.WithStack(err)
	}
	req.Header.Set("Client-Id", c.clientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return VOD{}, errors.Errorf("%v\n%s", err, string(dump))
	}
	defer resp.Body.Close()
	if s := resp.StatusCode; s < 200 || s >= 300 {
		return VOD{}, errors.Errorf("invalid status code %d\n%s", s, string(dump))
	}

	type respPayload struct {
		Data struct {
			Video struct {
				Title string `json:"title"`
			} `json:"video"`
		} `json:"data"`
	}

	var p respPayload
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return VOD{}, errors.Errorf("%v\n%s", err, string(dump))
	}
	return VOD{Title: p.Data.Video.Title}, nil
}
