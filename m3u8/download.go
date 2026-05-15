package m3u8

import (
	"context"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Download fetches and stitches together all segments from an HLS media playlist.
// r must contain the m3u8 content. baseURL is used to resolve relative segment
// URLs (pass the playlist's source URL when fetched over HTTP, or "" for a local
// file with absolute segment URLs). start and end can trim the output by time;
// pass zero values to download the full playlist.
// Segment HTTP requests are made lazily as the returned io.ReadCloser is read.
func Download(ctx context.Context, client *http.Client, r io.Reader, baseURL string, start, end time.Duration) (io.ReadCloser, error) {
	media, err := Media(r, baseURL)
	if err != nil {
		return nil, err
	}

	segments, err := sliceSegments(media.Segments, start, end)
	if err != nil {
		return nil, err
	}

	var fns []downloadFunc
	for _, segment := range segments {
		req, err := http.NewRequest(http.MethodGet, segment.URL, nil)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		req = req.WithContext(ctx)
		fns = append(fns, prepare(client, req))
	}

	return &merger{downloads: fns}, nil
}

func sliceSegments(segments []MediaSegment, start, end time.Duration) ([]MediaSegment, error) {
	if start < 0 || end < 0 {
		return nil, errors.New("negative timestamps are not allowed")
	}
	if start >= end && end != 0 {
		return nil, errors.New("end timestamp is not after start timestamp")
	}
	if start == 0 && end == 0 {
		return segments, nil
	}
	if end == 0 {
		end = time.Duration(math.MaxInt64)
	}
	var slice []MediaSegment
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
		for _, s := range segments {
			dur += s.Duration
		}
		return nil, errors.Errorf("timestamps are not a subset of the media (duration: %v)", dur)
	}
	return slice, nil
}

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

// merger stitches multiple downloadFuncs into a single io.ReadCloser.
type merger struct {
	downloads []downloadFunc
	index     int
	current   io.ReadCloser
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

func (r *merger) Read(p []byte) (int, error) {
	for {
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

func (r *merger) Close() error { return nil }

// Chunks returns the total number of segments.
func (r *merger) Chunks() int { return len(r.downloads) }

// Current returns the number of segments already processed.
func (r *merger) Current() int { return r.index }
