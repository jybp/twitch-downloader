package twitchdl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/jybp/twitch-downloader/m3u8"
	"github.com/jybp/twitch-downloader/twitch"
)

// Name return the name of the video: Channel Name - Video name.
func Name(ctx context.Context, client *http.Client, clientID, vURL string) (string, error) {
	api := twitch.New(client, clientID)
	id, vType, err := twitch.ID(vURL)
	if err != nil {
		return "", err
	}
	switch vType {
	case twitch.TypeVOD:
		vod, err := api.VOD(ctx, id)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s - %s", vod.Owner.DisplayName, vod.Title), nil
	case twitch.TypeClip:
		clip, err := api.Clip(ctx, id)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s - %s", clip.Broadcaster.DisplayName, clip.Title), nil
	default:
		return "", errors.Errorf("unsupported video type %d", vType)
	}
}

// Qualities return the qualities available.
func Qualities(ctx context.Context, client *http.Client, clientID, vURL string) ([]string, error) {
	api := twitch.New(client, clientID)
	id, vType, err := twitch.ID(vURL)
	if err != nil {
		return nil, err
	}
	switch vType {
	case twitch.TypeVOD:
		m3u8raw, err := api.M3U8(ctx, id)
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
	case twitch.TypeClip:
		clip, err := api.ClipVideo(ctx, id)
		if err != nil {
			return nil, err
		}
		var qualities []string
		for _, quality := range clip.Qualities {
			qualities = append(qualities, fmt.Sprintf(clipQualityFramerateFormat, quality.Quality, quality.FrameRate))
		}
		return qualities, nil
	default:
		return nil, errors.Errorf("unsupported video type %d", vType)
	}
}

// Download sets up the download of the VOD "vodId" with quality "quality"
// using the provided http.Client.
// The download is actually perfomed when the returned io.Reader is being read.
func Download(ctx context.Context, client *http.Client, clientID, vURL, quality string, start, end time.Duration) (io.ReadCloser, error) {
	id, vType, err := twitch.ID(vURL)
	if err != nil {
		return nil, err
	}
	switch vType {
	case twitch.TypeVOD:
		return downloadVOD(ctx, client, clientID, id, quality, start, end)
	case twitch.TypeClip:
		return downloadClip(ctx, client, clientID, id, quality)
	default:
		return nil, errors.Errorf("unsupported video type %d", vType)
	}
}
