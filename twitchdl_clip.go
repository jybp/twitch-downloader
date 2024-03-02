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

const clipQualityFramerateFormat = "%sp%.f"

func downloadClip(ctx context.Context, client *http.Client, clientID, id, quality string) (io.ReadCloser, error) {
	api := twitch.New(client, clientID)
	clip, err := api.ClipVideo(ctx, id)
	if err != nil {
		return nil, err
	}
	var dlURL string
	for _, q := range clip.Qualities {
		str := fmt.Sprintf(clipQualityFramerateFormat, q.Quality, q.FrameRate)
		if str != quality {
			continue
		}
		dlURL = fmt.Sprintf("%s?sig=%s&token=%s",
			q.SourceURL,
			url.QueryEscape(clip.Token.Signature),
			url.QueryEscape(clip.Token.Value))
	}
	if len(dlURL) == 0 {
		return nil, errors.Errorf("Quality %s not found", quality)
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
