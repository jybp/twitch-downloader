package twitchdl

import (
	"bytes"
	"context"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/jybp/twitch-downloader/m3u8"
	"github.com/jybp/twitch-downloader/twitch"
	"github.com/pkg/errors"
)

func downloadVOD(ctx context.Context, client *http.Client, clientID, id, quality string, start, end time.Duration) (io.ReadCloser, error) {
	api := twitch.New(client, clientID)
	m3u8raw, err := api.M3U8(ctx, id)
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
	segments, err := sliceSegments(media.Segments, start, end)
	if err != nil {
		return nil, err
	}
	for _, segment := range segments {
		req, err := http.NewRequest(http.MethodGet, segment.URL, nil)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		req = req.WithContext(ctx)
		downloadFns = append(downloadFns, prepare(client, req))
	}

	return &merger{downloads: downloadFns}, nil
}

func sliceSegments(segments []m3u8.MediaSegment, start, end time.Duration) ([]m3u8.MediaSegment, error) {
	if start < 0 || end < 0 {
		return nil, errors.New("Negative timestamps are not allowed")
	}
	if start >= end && end != time.Duration(0) {
		return nil, errors.New("End timestamp is not after Start timestamp")
	}
	if start == time.Duration(0) && end == time.Duration(0) {
		return segments, nil
	}
	if end == time.Duration(0) {
		end = time.Duration(math.MaxInt64)
	}
	slice := []m3u8.MediaSegment{}
	segmentStart := time.Duration(0)
	for _, segment := range segments {
		segmentEnd := segmentStart + segment.Duration
		if segmentEnd <= start {
			segmentStart += segment.Duration
			continue
		}
		if segmentStart >= end {
			break
		}
		slice = append(slice, segment)
		segmentStart += segment.Duration
	}
	if len(slice) == 0 {
		var dur time.Duration
		for _, segment := range segments {
			dur += segment.Duration
		}
		return nil, errors.Errorf("Timestamps are not a subset of the video (video duration is %v)", dur)
	}
	return slice, nil
}

// downloadFunc describes a func that peform an HTTP request and returns the response.Body
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

// merger merges the several downloadFunc into a single io.Reader.
type merger struct {
	downloads []downloadFunc

	index   int
	current io.ReadCloser
	err     error
}

func (r *merger) next() error {
	if r.index >= len(r.downloads) {
		r.current = nil
		r.index++
		return nil
	}
	var err error
	r.current, err = r.downloads[r.index]()
	r.index++
	return err
}

// Read allows merger to implement io.Reader.
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

func (r *merger) Close() error {
	return nil
}

// Chunks returns the number of chunks.
func (r *merger) Chunks() int {
	return len(r.downloads)
}

// Current returns the number of chunks already processed.
func (r *merger) Current() int {
	return r.index
}
