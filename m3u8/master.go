package m3u8

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// https://developer.apple.com/documentation/http_live_streaming

// Resolution describes the optimal pixel
// resolution at which to display all the video in the Variant
// Stream.
//
// https://tools.ietf.org/html/rfc8216#page-30
type Resolution struct {
	Width  int
	Height int
}

// Alternative describes altenative renditions of the same content.
//
// https://tools.ietf.org/html/rfc8216#page-25
type Alternative struct {
	// Required
	Type    string
	GroupID string
	Name    string
	// Optional
	Autoselect bool
	Default    bool
}

// Variant specifies a Variant Stream.
//
// https://tools.ietf.org/html/rfc8216#page-29
type Variant struct {
	// Required
	URL       string
	Bandwidth int
	// Optional
	Codecs       []string
	Resolution   Resolution
	Video        string
	Audio        string
	Alternatives []Alternative
}

// MasterPlaylist defines the Variant Streams, Renditions, and
// other global parameters of the presentation.
//
// https://tools.ietf.org/html/rfc8216#page-25
type MasterPlaylist struct {
	Variants []Variant
}

func attributes(line string) (map[string]string, error) {
	insideQuotes := false
	fn := func(c rune) bool {
		if '"' == c {
			insideQuotes = !insideQuotes
			return false
		}
		return ',' == c && !insideQuotes
	}
	list := strings.FieldsFunc(line, fn)
	attr := map[string]string{}
	for _, it := range list {
		kv := strings.Split(it, "=")
		if len(kv) != 2 {
			return attr, errors.New("malformed attribute")
		}
		attr[kv[0]] = strings.Trim(kv[1], `"`)
	}
	return attr, nil
}

// Master parses a Master Playlist.
func Master(r io.Reader) (MasterPlaylist, error) {
	scanner := bufio.NewScanner(r)

	if !scanner.Scan() {
		return MasterPlaylist{}, errors.WithStack(io.ErrUnexpectedEOF)
	}
	sig := scanner.Text()
	if sig != "#EXTM3U" {
		return MasterPlaylist{}, errors.New("invalid signature")
	}

	playlist := MasterPlaylist{}
	alternatives := []Alternative{}
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			attr, err := attributes(line[18:])
			if err != nil {
				return playlist, err
			}
			variant := Variant{}

			bandwidth, ok := attr["BANDWIDTH"]
			if ok {
				bandwidth, err := strconv.Atoi(bandwidth)
				if err != nil {
					return playlist, errors.WithStack(err)
				}
				variant.Bandwidth = bandwidth
			}

			codecs, _ := attr["CODECS"]
			variant.Codecs = strings.Split(codecs, ",")

			resolution, _ := attr["RESOLUTION"]
			fmt.Sscanf(resolution, "%dx%d", &variant.Resolution.Width, &variant.Resolution.Height)

			variant.Video, _ = attr["VIDEO"]
			variant.Audio, _ = attr["AUDIO"]

			if !scanner.Scan() {
				return playlist, errors.WithStack(io.ErrUnexpectedEOF)
			}
			variant.URL = scanner.Text()

			playlist.Variants = append(playlist.Variants, variant)
			continue
		}

		if strings.HasPrefix(line, "#EXT-X-MEDIA:") {
			attr, err := attributes(line[13:])
			if err != nil {
				return playlist, err
			}
			alternative := Alternative{}

			alternative.Type, _ = attr["TYPE"]
			alternative.GroupID, _ = attr["GROUP-ID"]
			alternative.Name, _ = attr["NAME"]

			autoselect, _ := attr["AUTOSELECT"]
			alternative.Autoselect = autoselect == "YES"

			def, _ := attr["DEFAULT"]
			alternative.Default = def == "YES"

			alternatives = append(alternatives, alternative)
			continue
		}

		// Discard line.
	}

	if err := scanner.Err(); err != nil {
		return playlist, errors.WithStack(err)
	}

	// Match alternatives with variants.
	for _, alt := range alternatives {
		var match func(v Variant, a Alternative) bool
		if alt.Type == "VIDEO" {
			match = func(v Variant, a Alternative) bool { return a.GroupID == v.Video }
		} else if alt.Type == "AUDIO" {
			match = func(v Variant, a Alternative) bool { return a.GroupID == v.Audio }
		} else {
			return playlist, errors.Errorf("unsupported alternative type %s", alt.Type)
		}

		for i, v := range playlist.Variants {
			if match(v, alt) {
				playlist.Variants[i].Alternatives = append(playlist.Variants[i].Alternatives, alt)
				break
			}
		}
	}

	return playlist, nil
}
