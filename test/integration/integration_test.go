package integration

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/jybp/twitch-downloader/twitch"

	twitchdl "github.com/jybp/twitch-downloader"
)

var (
	clientID  string
	vodID     string
	quality   string
	nocleanup bool

	skip bool

	once sync.Once
)

func parseFlags() {
	once.Do(func() {
		flag.StringVar(&clientID, "clientID", "", "clientID")
		flag.StringVar(&vodID, "vodID", "", "vodID")
		flag.StringVar(&quality, "quality", "", "quality")
		flag.BoolVar(&nocleanup, "nocleanup", false, "do not clean up tmp file")
		flag.Parse()
		if len(clientID) == 0 || len(vodID) == 0 || len(quality) == 0 {
			skip = true
		}
	})
}

type transport struct {
	*testing.T
}

func (t transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if testing.Verbose() {
		print(req.Method, ": ", req.URL.String(), "\n")
	}
	return http.DefaultTransport.RoundTrip(req)
}

func client(t *testing.T) *http.Client {
	return &http.Client{Transport: transport{t}}
}

func TestVOD(t *testing.T) {
	parseFlags()
	if skip {
		t.SkipNow()
	}
	api := twitch.New(client(t), clientID)
	vod, err := api.VOD(context.Background(), vodID)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if testing.Verbose() {
		t.Logf("%v", vod.Title)
	}
}

func TestQualities(t *testing.T) {
	parseFlags()
	if skip {
		t.SkipNow()
	}
	qualities, err := twitchdl.Qualities(context.Background(), client(t), clientID, vodID)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if testing.Verbose() {
		t.Logf("%v", qualities)
	}
	if len(qualities) == 0 {
		t.Fatal("0 qualities")
	}
}

// The test can take a long time to complete. Make sure to use the -timeout flag.
func TestDownload(t *testing.T) {
	parseFlags()
	if skip {
		t.SkipNow()
	}

	reader, err := twitchdl.Download(context.Background(), client(t), clientID, vodID, quality, 0, 0)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	tmpFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("twitchdl_%s_%s_*.mp4", vodID, quality))
	if err != nil {
		t.Fatalf("cannot create temporary file: %+v", err)
	}
	if !nocleanup {
		defer os.Remove(tmpFile.Name())
	}
	if _, err := io.Copy(tmpFile, reader); err != nil {
		t.Fatalf("%+v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("%+v", err)
	}
	if testing.Verbose() {
		t.Logf("%s", tmpFile.Name())
	}
}
