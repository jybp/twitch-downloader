package twitchdl

import (
	"bytes"
	"context"
	"io"
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

	return m3u8.Download(ctx, client, mediaResp.Body, variant.URL, start, end)
}
