package integration

import (
	"context"
	"flag"
	"net/http"
	"testing"

	"github.com/jybp/twitch-downloader/twitch"
)

var (
	clientID string
	vodID    string
	skip     bool
)

func init() {
	flag.StringVar(&clientID, "clientID", "", "clientID")
	flag.StringVar(&vodID, "vodID", "", "vodID")
	flag.Parse()
	if len(clientID) == 0 || len(vodID) == 0 {
		skip = true
	}
}

func TestM3U(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	client := twitch.New(http.DefaultClient, clientID)
	m3u, err := client.VODM3U8(context.Background(), vodID)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if testing.Verbose() {
		t.Logf("%s", string(m3u))
	}
}
