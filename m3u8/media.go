package m3u8

import (
	"bufio"
	"io"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// MediaSegment describes a chunk.
//
// https://tools.ietf.org/html/rfc8216#page-6
type MediaSegment struct {
	Number   int
	Duration time.Duration
	URL      string
}

// MediaPlaylist contains a series of Media Segments that make up the
// overall presentation.
//
// https://tools.ietf.org/html/rfc8216#page-22
type MediaPlaylist struct {
	TargetDuration time.Duration
	Type           string
	Sequence       int
	Ended          bool
	Segments       []MediaSegment
}

// Media parses a Media Playlist.
// URL is an optional argument that matches the Variant URL inside the Master Playlist.
// It is used to construct full URLs if the URLs inside Media Segments are relative.
func Media(r io.Reader, URL string) (MediaPlaylist, error) {
	var baseURL *url.URL
	if len(URL) > 0 {
		var err error
		baseURL, err = url.Parse(strings.TrimRight(URL, path.Base(URL)))
		if err != nil {
			return MediaPlaylist{}, errors.WithStack(err)
		}
	}

	scanner := bufio.NewScanner(r)

	if !scanner.Scan() {
		return MediaPlaylist{}, errors.WithStack(io.ErrUnexpectedEOF)
	}
	sig := scanner.Text()
	if sig != "#EXTM3U" {
		return MediaPlaylist{}, errors.New("invalid signature")
	}

	playlist := MediaPlaylist{}
	var segmentIndex = 0
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#EXT-X-TARGETDURATION:") {
			d, err := strconv.Atoi(line[22:])
			if err != nil {
				return playlist, errors.WithStack(err)
			}
			playlist.TargetDuration = time.Second * time.Duration(d)
			continue
		}

		if strings.HasPrefix(line, "#EXT-X-PLAYLIST-TYPE:") {
			playlist.Type = line[21:]
			continue
		}

		if strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:") {
			n, err := strconv.Atoi(line[22:])
			if err != nil {
				return playlist, errors.WithStack(err)
			}
			playlist.Sequence = n
			continue
		}

		if strings.HasPrefix(line, "#EXTINF:") {
			segment := MediaSegment{}
			segment.Number = playlist.Sequence + segmentIndex
			segmentIndex++

			firstComma := strings.Index(line, ",")
			if firstComma == -1 {
				firstComma = len(line)
			}
			d, err := strconv.ParseFloat(line[8:firstComma], 64)
			if err != nil {
				return playlist, errors.WithStack(err)
			}
			segment.Duration = time.Duration(d * float64(time.Second))

			if !scanner.Scan() {
				return playlist, errors.WithStack(io.ErrUnexpectedEOF)
			}
			segment.URL = scanner.Text()
			if baseURL != nil {
				segmentURL, err := url.Parse(segment.URL)
				if err != nil {
					return playlist, errors.WithStack(err)
				}
				if !segmentURL.IsAbs() {
					segment.URL = baseURL.ResolveReference(segmentURL).String()
				}
			}

			playlist.Segments = append(playlist.Segments, segment)
			continue
		}

		if line == "#EXT-X-ENDLIST" {
			playlist.Ended = true
			continue
		}

		// Discard line.
	}

	return playlist, errors.WithStack(scanner.Err())
}
