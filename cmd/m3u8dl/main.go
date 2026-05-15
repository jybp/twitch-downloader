package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jybp/twitch-downloader/m3u8"
	"github.com/pkg/errors"
)

var input, output string
var start, end time.Duration
var verbose bool

func init() {
	log.SetFlags(0)

	flag.StringVar(&input, "i", "", "URL or file path of the m3u8 media playlist to download.")
	flag.StringVar(&output, "o", "out.ts", "Output file path. (default: out.ts)")
	flag.DurationVar(&start, "start", 0, "Start offset. Example: 1h23m45s (optional)")
	flag.DurationVar(&end, "end", 0, "End offset. Example: 1h34m56s (optional)")
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
	if len(input) == 0 {
		flag.PrintDefaults()
		return nil
	}

	var r io.ReadCloser
	var baseURL string

	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		resp, err := http.Get(input)
		if err != nil {
			return errors.Wrapf(err, "fetching %s failed", input)
		}
		defer resp.Body.Close()
		if s := resp.StatusCode; s < 200 || s >= 300 {
			return errors.Errorf("fetching %s: HTTP %d", input, s)
		}
		r = resp.Body
		baseURL = input
	} else {
		f, err := os.Open(input)
		if err != nil {
			return errors.Wrapf(err, "opening %s failed", input)
		}
		defer f.Close()
		r = f
		baseURL = ""
	}

	download, err := m3u8.Download(context.Background(), http.DefaultClient, r, baseURL, start, end)
	if err != nil {
		return errors.Wrapf(err, "preparing download from %s failed", input)
	}

	f, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return errors.Wrapf(err, "cannot create file %s", output)
	}

	fmt.Printf("Downloading: %s\n", f.Name())

	if _, err := io.Copy(f, &reader{r: download}); err != nil {
		return errors.Wrapf(err, "writing to file %s failed", output)
	}
	if err := f.Close(); err != nil {
		return errors.Wrapf(err, "closing file %s failed", output)
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
	if r.from.IsZero() {
		return 0
	}
	elapsed := uint64(time.Since(r.from))
	if elapsed == 0 {
		return 0
	}
	return uint64(r.n) * uint64(time.Second) / elapsed
}
