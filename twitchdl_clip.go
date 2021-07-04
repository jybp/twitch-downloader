package twitchdl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/jybp/twitch-downloader/twitch"
	"github.com/pkg/errors"
)

const clipQualityFramerateFormat = "%sp%d"

func downloadClip(ctx context.Context, client *http.Client, clientID, id, targetQuality string) (io.ReadCloser, error) {
	api := twitch.New(client, clientID)
	clip, err := api.ClipVideo(ctx, id)
	if err != nil {
		return nil, err
	}
	var dlURL string
	for _, quality := range clip.Qualities {
		str := fmt.Sprintf(clipQualityFramerateFormat, quality.Quality, quality.FrameRate)
		if str != targetQuality {
			continue
		}
		dlURL = fmt.Sprintf("%s?sig=%s&token=%s",
			quality.SourceURL,
			url.QueryEscape(clip.Token.Signature),
			url.QueryEscape(clip.Token.Value))
	}
	if len(dlURL) == 0 {
		return nil, errors.Errorf("Quality %s not found", targetQuality)
	}
	req, err := http.NewRequest(http.MethodGet, dlURL, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if s := resp.StatusCode; s < 200 || s >= 300 {
		return nil, errors.Errorf("%d: %s", s, req.URL)
	}
	return resp.Body, nil
}
