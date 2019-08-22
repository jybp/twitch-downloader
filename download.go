package twitchdl

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/jybp/twitch-downloader/m3u8"
	"github.com/jybp/twitch-downloader/twitch"
	"github.com/pkg/errors"
)

// Qualities return the qualities available for the VOD "vodID".
func Qualities(ctx context.Context, client *http.Client, clientID, vodID string) ([]string, error) {
	api := twitch.New(client, clientID)
	m3u8raw, err := api.M3U8(ctx, vodID)
	if err != nil {
		return nil, err
	}
	master, err := m3u8.Master(bytes.NewReader(m3u8raw))
	if err != nil {
		return nil, err
	}
	var qualities []string
	for _, variant := range master.Variants {
		for _, alt := range variant.Alternatives {
			qualities = append(qualities, alt.Name)
		}
	}
	return qualities, nil
}

// Download sets up the download of the VOD "vodId" with quality "quality"
// using the provided http.Client.
// The download is actually perfomed when the returned io.Reader is being read.
// "size" is an rough estimate of the expected file size
func Download(ctx context.Context, client *http.Client, clientID, vodID, quality string) (r io.Reader, err error) {
	api := twitch.New(client, clientID)
	m3u8raw, err := api.M3U8(ctx, vodID)
	if err != nil {
		return nil, err
	}
	master, err := m3u8.Master(bytes.NewReader(m3u8raw))
	if err != nil {
		return nil, err
	}

	var variant m3u8.Variant
L:
	for _, v := range master.Variants {
		for _, alt := range v.Alternatives {
			if alt.Name != quality {
				continue
			}
			variant = v
			break L
		}
	}

	if len(variant.URL) == 0 {
		return nil, errors.Errorf("quality %s not found", quality)
	}

	mediaResp, err := client.Get(variant.URL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer mediaResp.Body.Close()
	media, err := m3u8.Media(mediaResp.Body, variant.URL)
	if err != nil {
		return nil, err
	}

	var downloadFns []downloadFunc
	for _, segment := range media.Segments {
		req, err := http.NewRequest(http.MethodGet, segment.URL, nil)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		downloadFns = append(downloadFns, prepare(client, req))
	}

	return &merger{downloads: downloadFns}, nil
}

// dlFunc describes a func that peform an HTTP request and returns the response.Body
type downloadFunc func() (io.ReadCloser, error)

func prepare(client *http.Client, req *http.Request) downloadFunc {
	return func() (io.ReadCloser, error) {
		resp, err := client.Do(req)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if s := resp.StatusCode; s < 200 || s >= 300 {
			return nil, errors.Errorf("%d: %s", s, req.URL)
		}
		return resp.Body, nil
	}
}

// merger merges the "downloads" into a single io.Reader.
type merger struct {
	downloads []downloadFunc

	index   int
	current io.ReadCloser
	err     error
}

func (r *merger) next() error {
	if r.index >= len(r.downloads) {
		r.current = nil
		return nil
	}
	var err error
	r.current, err = r.downloads[r.index]()
	r.index++
	return err
}

func (r *merger) Read(p []byte) (int, error) {
	for {
		if r.err != nil {
			return 0, r.err
		}
		if r.current != nil {
			n, err := r.current.Read(p)
			if err == io.EOF {
				err = r.current.Close()
				r.current = nil
			}
			return n, errors.WithStack(err)
		}
		if err := r.next(); err != nil {
			return 0, err
		}
		if r.current == nil {
			return 0, io.EOF
		}
	}
}
