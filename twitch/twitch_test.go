package twitch_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jybp/twitch-downloader/twitch"
)

func setup(t *testing.T, cliendID string) (client twitch.Client, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	url := server.URL + "/"
	client = twitch.Custom(http.DefaultClient, cliendID, url, url)
	return client, mux, server.Close
}

func TestVODM3U8(t *testing.T) {
	client, mux, teardown := setup(t, "client_id")
	defer teardown()

	mux.HandleFunc("/api/vods/123/access_token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("expected GET method, got: %s", r.Method)
		}
		assert.Equal(t, "client_id", r.URL.Query().Get("client_id"))
		fmt.Fprintf(w, `{"token":"token","sig":"sig"}`)
	})
	mux.HandleFunc("/vod/123", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("expected GET method, got: %s", r.Method)
		}
		assert.Equal(t, "token", r.URL.Query().Get("nauth"))
		assert.Equal(t, "sig", r.URL.Query().Get("nauthsig"))
		assert.Equal(t, "true", r.URL.Query().Get("allow_audio_only"))
		assert.Equal(t, "true", r.URL.Query().Get("allow_source"))
		fmt.Fprintf(w, "#EXTM3U\n")
	})

	m3u, err := client.M3U8(context.Background(), "123")
	if err != nil {
		t.Fatalf("%+v", err)
	}
	assert.Equal(t, "#EXTM3U\n", string(m3u))
	if testing.Verbose() {
		t.Logf("%s", string(m3u))
	}
}
