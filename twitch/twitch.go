package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// https://dev.twitch.tv/docs/api/reference/
// https://github.com/videolan/vlc/blob/0b018b348f47cda82863809ab0385cb993c8aa33/share/lua/playlist/twitch.lua#L81

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
		return token{}, errors.Errorf("%d\n%s\n%s", s, u, string(b))
	}
	var t token
	return t, errors.WithStack(json.NewDecoder(resp.Body).Decode(&t))
}

// M3U8 retrieves the M3U8 file of a specific VOD.
func (c *Client) M3U8(ctx context.Context, id string) ([]byte, error) {
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
		return nil, errors.Errorf("%d\n%s\n%s", s, u, string(b))
	}
	return ioutil.ReadAll(resp.Body)
}

// VOD describes a twitch VOD.
type VOD struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	UserName     string    `json:"user_name"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	PublishedAt  time.Time `json:"published_at"`
	URL          string    `json:"url"`
	ThumbnailURL string    `json:"thumbnail_url"`
	Viewable     string    `json:"viewable"`
	ViewCount    int       `json:"view_count"`
	Language     string    `json:"language"`
	Type         string    `json:"type"`
	Duration     string    `json:"duration"`
}

type data struct {
	Data       []VOD `json:"data"`
	Pagination struct {
		Cursor string `json:"cursor"`
	} `json:"pagination"`
}

// VOD retrieves the video informations of a specific VOD.
func (c *Client) VOD(ctx context.Context, id string) (VOD, error) {
	u := fmt.Sprintf("%shelix/videos?id=%s", c.apiURL, id)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return VOD{}, errors.WithStack(err)
	}
	req.Header.Add("Client-ID", c.clientID)
	resp, err := c.client.Do(req)
	if err != nil {
		return VOD{}, errors.WithStack(err)
	}
	defer resp.Body.Close()
	if s := resp.StatusCode; s < 200 || s >= 300 {
		b, _ := ioutil.ReadAll(resp.Body)
		return VOD{}, errors.Errorf("%d\n%s\n%s", s, u, string(b))
	}
	var d data
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return VOD{}, errors.WithStack(err)
	}
	if len(d.Data) != 1 {
		return VOD{}, errors.New("unexpected data length")
	}
	return d.Data[0], nil
}
