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
	"github.com/pkg/errors"
)

// Must be injected at build time using the -ldflags flag.
// go build -ldflags "-X main.defaultClientID=YourClientID"
var defaultClientID string

// Flags
var clientID, url, quality, output string
var start, end time.Duration
var verbose bool

func init() {
	log.SetFlags(0)

	flag.StringVar(&url, "url", "", `The URL of the twitch VOD or Clip to download.`)
	flag.StringVar(&quality, "q", "", "Quality of the video to download. Omit this flag to print the available qualities.\nUse \"best\" to automatically select the highest quality.")
	flag.StringVar(&output, "o", "", "Path where the video will be downloaded. Example: `-o my-video.ts`. (optional)")
	flag.DurationVar(&start, "start", time.Duration(0), "Specify \"start\" to download a subset of the VOD. Example: 1h23m45s (optional)")
	flag.DurationVar(&end, "end", time.Duration(0), "Specify \"end\" to download a subset of the VOD. Example: 1h34m56s (optional)")
	flag.StringVar(&clientID, "client-id", "", "Use a specific twitch.tv API client ID. (optional)")
	flag.BoolVar(&verbose, "v", false, "Verbose errors. (optional)")
	flag.Parse()
}

func main() {
	errVerb := "%v"
	if verbose {
		errVerb = "%+v"
	}
	if err := run(); err != nil {
		log.Fatalf(errVerb, err)
	}
}

func run() error {
	if len(clientID) > 0 {
		defaultClientID = clientID
	}

	if len(defaultClientID) == 0 {
		panic("no default client id specified")
	}

	if len(url) == 0 {
		flag.PrintDefaults()
		return nil
	}

	name, err := twitchdl.Name(context.Background(), http.DefaultClient, defaultClientID, url)
	if err != nil {
		return errors.Wrapf(err, "Retrieving name for URL %s failed", url)
	}

	if len(quality) == 0 {
		qualities, err := twitchdl.Qualities(context.Background(), http.DefaultClient, defaultClientID, url)
		if err != nil {
			return errors.Wrapf(err, "Retrieving qualities for URL %s failed", url)
		}
		fmt.Printf("%s\n%s\n", name, strings.Join(qualities, "\n"))
		return nil
	}
	if quality == "best" {
		qualities, err := twitchdl.Qualities(context.Background(), http.DefaultClient, defaultClientID, url)
		if err != nil || len(qualities) == 0 {
			return errors.Wrapf(err, "Retrieving qualities for URL %s failed", url)
		}
		quality = qualities[0]
	}

	download, err := twitchdl.Download(context.Background(), http.DefaultClient, defaultClientID, url, quality, start, end)
	if err != nil {
		return errors.Wrapf(err, "Retrieving stream for URL %s failed", url)
	}

	path, filename := filepath.Split(output)
	if len(filename) == 0 {
		ext := "ts"
		if strings.Contains(strings.ToLower(quality), "audio") {
			ext = "aac"
		}
		filename = fmt.Sprintf("%s (%s).%s", name, quality, ext)
	}
	output = filepath.Join(path, filename)

	f, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return errors.Wrapf(err, "Cannot create file %s", output)
	}

	fmt.Printf("Downloading: %s\n", f.Name())

	if _, err := io.Copy(f, &reader{r: download}); err != nil {
		return errors.Wrapf(err, "Writing to file %s failed", output)
	}
	if err := f.Close(); err != nil {
		return errors.Wrapf(err, "Closing file %s failed", output)
	}
	fmt.Printf("\rDone%-25s\n", " ")
	return nil
}

// reader prints the download progress every second.
type reader struct {
	r io.Reader

	from time.Time
	n    uint64
	t    uint64
}

func (r *reader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	r.n += uint64(n)
	r.t += uint64(n)
	if time.Since(r.from) > time.Second {
		fmt.Printf("\r%-12s %-10s",
			r.btos(r.bitrate())+"/s",
			r.btos(r.t))
		r.from = time.Now()
		r.n = 0
	}
	return
}

func (*reader) btos(b uint64) string {
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

func (r *reader) bitrate() uint64 {
	return uint64(r.n) * uint64(time.Second) / uint64(time.Since(r.from))
}
