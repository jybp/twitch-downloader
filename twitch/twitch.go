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

type VideoType int

const (
	TypeVOD VideoType = iota
	TypeClip
)

// ID extract the ID/slug and type from a VOD url.
func ID(URL string) (string, VideoType, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return "", 0, errors.WithStack(err)
	}
	if !strings.Contains(u.Hostname(), "twitch.tv") {
		return "", 0, errors.Errorf("URL host for %s is not twitch.tv", URL)
	}
	if strings.HasPrefix(u.Path, "/videos/") || strings.Contains(u.Path, "/video/") {
		_, id := path.Split(u.Path)
		return id, TypeVOD, nil
	}
	if strings.Contains(u.Path, "/clip/") || strings.Contains(u.Hostname(), "clips.twitch.tv") {
		_, id := path.Split(u.Path)
		return id, TypeClip, nil
	}
	return "", 0, errors.New("Cannot extract VOD ID or clip slug from URL")
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
	return Client{client, clientID, "https://gql.twitch.tv/gql", "https://usher.ttvnw.net"}
}

// Custom returns a new twitch API client with custom API endpoints
func Custom(client *http.Client, clientID, apiURL, usherAPIURL string) Client {
	return Client{client, clientID, apiURL, usherAPIURL}
}

func (c *Client) vodToken(ctx context.Context, id string) (token, sig string, _ error) {
	gqlPayload := `{"operationName":"PlaybackAccessToken_Template","query":"query PlaybackAccessToken_Template($login: String!, $isLive: Boolean!, $vodID: ID!, $isVod: Boolean!, $playerType: String!) {  streamPlaybackAccessToken(channelName: $login, params: {platform: \"web\", playerBackend: \"mediaplayer\", playerType: $playerType}) @include(if: $isLive) {    value    signature    __typename  }  videoPlaybackAccessToken(id: $vodID, params: {platform: \"web\", playerBackend: \"mediaplayer\", playerType: $playerType}) @include(if: $isVod) {    value    signature    __typename  }}","variables":{"isLive":false,"login":"","isVod":true,"vodID":"%s","playerType":"site"}}`
	body := strings.NewReader(fmt.Sprintf(gqlPayload, id))
	req, err := http.NewRequest(http.MethodPost, c.apiURL, body)
	if err != nil {
		return "", "", errors.WithStack(err)
	}
	req = req.WithContext(ctx)
	req.Header.Set("Client-ID", c.clientID)
	type payload struct {
		Data struct {
			VideoPlaybackAccessToken struct {
				Value     string `json:"value"`
				Signature string `json:"signature"`
			} `json:"videoPlaybackAccessToken"`
		} `json:"data"`
	}
	var p payload
	if err := c.do(ctx, req, &p); err != nil {
		return "", "", err
	}
	return p.Data.VideoPlaybackAccessToken.Value, p.Data.VideoPlaybackAccessToken.Signature, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) error {
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return errors.WithStack(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Errorf("%v\n%s", err, string(dump))
	}
	defer resp.Body.Close()
	if s := resp.StatusCode; s < 200 || s >= 300 {
		return errors.Errorf("invalid status code %d\n%s", s, string(dump))
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return errors.Errorf("%v\n%s", err, string(dump))
	}
	return nil
}

// VOD contains infos on a twitch VOD.
type VOD struct {
	Title string `json:"title"`
	Owner struct {
		DisplayName string `json:"displayName"`
	} `json:"owner"`
	Game struct {
		Name string `json:"name"`
	} `json:"game"`
}

// Name retrieves the name of the video from a URL.
func (c *Client) VOD(ctx context.Context, id string) (VOD, error) {
	gqlPayload := `{"operationName":"VideoMetadata","variables":{"channelLogin":"","videoID":"%s"},"extensions":{"persistedQuery":{"version":1,"sha256Hash":"226edb3e692509f727fd56821f5653c05740242c82b0388883e0c0e75dcbf687"}}}`
	body := strings.NewReader(fmt.Sprintf(gqlPayload, id))
	req, err := http.NewRequest(http.MethodPost, c.apiURL, body)
	if err != nil {
		return VOD{}, errors.WithStack(err)
	}
	req = req.WithContext(ctx)
	req.Header.Set("Client-Id", c.clientID)
	type payload struct {
		Data struct {
			Video VOD `json:"video"`
		} `json:"data"`
	}
	var p payload
	if err := c.do(ctx, req, &p); err != nil {
		return VOD{}, err
	}
	return p.Data.Video, nil
}

type ClipVideo struct {
	Token struct {
		Signature string `json:"signature"`
		Value     string `json:"value"`
	} `json:"playbackAccessToken"`
	Qualities []struct {
		FrameRate float64 `json:"frameRate"`
		Quality   string  `json:"quality"`
		SourceURL string  `json:"sourceURL"`
	} `json:"videoQualities"`
}

func (c *Client) ClipVideo(ctx context.Context, slug string) (ClipVideo, error) {
	gqlPayload := `{"operationName":"VideoAccessToken_Clip","variables":{"slug":"%s"},"extensions":{"persistedQuery":{"version":1,"sha256Hash":"36b89d2507fce29e5ca551df756d27c1cfe079e2609642b4390aa4c35796eb11"}}}`
	body := strings.NewReader(fmt.Sprintf(gqlPayload, slug))
	req, err := http.NewRequest(http.MethodPost, c.apiURL, body)
	if err != nil {
		return ClipVideo{}, errors.WithStack(err)
	}
	req = req.WithContext(ctx)
	req.Header.Set("Client-Id", c.clientID)
	type payload struct {
		Data struct {
			ClipVideo ClipVideo `json:"clip"`
		} `json:"data"`
	}
	var p payload
	if err := c.do(ctx, req, &p); err != nil {
		return ClipVideo{}, err
	}
	return p.Data.ClipVideo, nil
}

type Clip struct {
	Broadcaster struct {
		DisplayName string `json:"displayName"`
	} `json:"broadcaster"`
	Title string `json:"title"`
	Game  struct {
		Name string `json:"name"`
	} `json:"game"`
}

func (c *Client) Clip(ctx context.Context, slug string) (Clip, error) {
	gqlPayload := `{"operationName":"ComscoreStreamingQuery","variables":{"channel":"","clipSlug":"%s","isClip":true,"isLive":false,"isVodOrCollection":false,"vodID":""},"extensions":{"persistedQuery":{"version":1,"sha256Hash":"e1edae8122517d013405f237ffcc124515dc6ded82480a88daef69c83b53ac01"}}}`
	body := strings.NewReader(fmt.Sprintf(gqlPayload, slug))
	req, err := http.NewRequest(http.MethodPost, c.apiURL, body)
	if err != nil {
		return Clip{}, errors.WithStack(err)
	}
	req = req.WithContext(ctx)
	req.Header.Set("Client-Id", c.clientID)
	type payload struct {
		Data struct {
			Clip Clip `json:"clip"`
		} `json:"data"`
	}
	var p payload
	if err := c.do(ctx, req, &p); err != nil {
		return Clip{}, err
	}
	return p.Data.Clip, nil
}

// M3U8 retrieves the M3U8 file of a specific VOD.
func (c *Client) M3U8(ctx context.Context, id string) ([]byte, error) {
	tok, sig, err := c.vodToken(ctx, id)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("%s/vod/%s?nauth=%s&nauthsig=%s&allow_audio_only=true&allow_source=true",
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
