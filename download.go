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
	m3u8raw, err := api.VODM3U8(ctx, vodID)
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
func Download(ctx context.Context, client *http.Client, clientID, vodID, quality string) (io.Reader, error) {
	api := twitch.New(client, clientID)
	m3u8raw, err := api.VODM3U8(ctx, vodID)
	if err != nil {
		return nil, err
	}
	master, err := m3u8.Master(bytes.NewReader(m3u8raw))
	if err != nil {
		return nil, err
	}

	var mediaURL string
L:
	for _, variant := range master.Variants {
		for _, alt := range variant.Alternatives {
			if alt.Name != quality {
				continue
			}
			mediaURL = variant.URL
			break L
		}
	}

	if len(mediaURL) == 0 {
		return nil, errors.Errorf("quality %s not found", quality)
	}

	mediaResp, err := client.Get(mediaURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer mediaResp.Body.Close()
	media, err := m3u8.Media(mediaResp.Body, mediaURL)
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
