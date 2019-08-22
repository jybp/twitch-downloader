package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	twitchdl "github.com/jybp/twitch-downloader"
	"github.com/jybp/twitch-downloader/twitch"
)

// Must be injected at build time using the -ldflags flag.
// go build -ldflags "-X main.defaultClientID=YourClientID"
var defaultClientID string

// Flags
var (
	clientID, vodID, quality, output string
	info                             bool
)

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	flag.StringVar(&clientID, "client-id", "", "Use a specific twitch.tv API client ID. Usage is optional.")
	flag.BoolVar(&info, "i", false, "Print available qualities.")
	flag.StringVar(&output, "o", "", `Path where the VOD will be downloaded. Usage is optional.
Specifying a filename is optional. 
Must not be present with the -info flag.`)
	flag.StringVar(&quality, "q", "", "Quality of the VOD to download. Must not be present with the -info flag.")
	flag.StringVar(&vodID, "id", "", `The ID of the VOD to download.
Can be inferred from the URL:
https://www.twitch.tv/videos/123 is the VOD with ID "123".`)
	flag.Parse()
}

func main() {
	if len(defaultClientID) == 0 {
		panic("no default client id specified")
	}

	if len(vodID) == 0 || (len(quality) > 0 == info) {
		flag.PrintDefaults()
		return
	}

	if len(clientID) == 0 {
		clientID = defaultClientID
	}

	api := twitch.New(http.DefaultClient, clientID)
	vod, err := api.VOD(context.Background(), vodID)
	if err != nil {
		log.Fatalf("Retrieving video informations for VOD %s failed: %v", vodID, err)
	}

	if info {
		qualities, err := twitchdl.Qualities(context.Background(), http.DefaultClient, clientID, vodID)
		if err != nil {
			log.Fatalf("Retrieving qualities for VOD %s failed: %v", vodID, err)
		}
		log.Printf("%s\n%s\n", vod.Title, strings.Join(qualities, "\n"))
		return
	}

	download, err := twitchdl.Download(context.Background(), http.DefaultClient, clientID, vodID, quality)
	if err != nil {
		log.Fatalf("Retrieving stream for VOD %s failed: %v", vodID, err)
	}

	ext := "mp4"
	if strings.Contains(strings.ToLower(quality), "audio") {
		ext = "mp4a"
	}
	path, filename := filepath.Split(output)
	if len(filename) == 0 {
		filename = fmt.Sprintf("%s (%s).%s", vod.Title, quality, ext)
	}
	output = filepath.Join(path, filename)

	f, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		log.Fatalf("Cannot create file %s: %v", output, err)
	}

	log.Printf("Downloading to: %s\n", f.Name())

	if _, err := io.Copy(f, &reader{r: download}); err != nil {
		log.Fatalf("Writing to file %s failed: %v", output, err)
	}
	if err := f.Close(); err != nil {
		log.Fatalf("Closing file %s failed: %v", output, err)
	}

	fmt.Fprintf(os.Stderr, "\rSuccessfully downloaded to: %s\n", f.Name())
}

// reader prints download progress every second.
type reader struct {
	r io.Reader

	s time.Time
	l time.Time
	n int
	t int
}

func (r *reader) Read(p []byte) (n int, err error) {
	if r.t == 0 {
		r.s = time.Now()
	}
	n, err = r.r.Read(p)
	r.n += n
	r.t += n
	if time.Now().Sub(r.l) > time.Second {
		fmt.Fprintf(os.Stderr, "\r%s/s (%s)", r.bps(), r.btos(int64(r.t)))
		r.l = time.Now()
		r.n = 0
	}
	return
}

func (*reader) btos(b int64) string {
	const u = 1024
	if b < u {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(u), 0
	for n := b / u; n >= u; n /= u {
		div *= u
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func (r *reader) bps() string {
	d := time.Now().Sub(r.l)
	bps := int64(r.n) * int64(time.Second) / int64(d)
	return r.btos(bps)
}
