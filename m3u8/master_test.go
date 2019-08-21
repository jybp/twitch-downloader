package m3u8_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jybp/twitch-downloader/m3u8"
)

func TestMaster(t *testing.T) {
	b := []byte(`#EXTM3U
#EXT-X-EXAMPLE-INFO:ORIGIN="origin"
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="chunked",NAME="1080p",AUTOSELECT=YES,DEFAULT=YES
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=6847192,CODECS="avc1.42C028,mp4a.40.2",RESOLUTION="1920x1080",VIDEO="chunked"
http://example.com/chunked/index-dvr.m3u8
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="720p30",NAME="720p"
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=2303475,CODECS="avc1.4D401F,mp4a.40.2",RESOLUTION="1280x720",VIDEO="720p30"
http://example.com/720p30/index-dvr.m3u8`)
	playlist, err := m3u8.Master(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if testing.Verbose() {
		t.Logf("%+v", playlist)
	}

	assert.Equal(t, 2, len(playlist.Variants))

	assert.Equal(t, "http://example.com/chunked/index-dvr.m3u8", playlist.Variants[0].URL)
	assert.Equal(t, 6847192, playlist.Variants[0].Bandwidth)
	assert.Equal(t, 2, len(playlist.Variants[0].Codecs))
	assert.Equal(t, "avc1.42C028", playlist.Variants[0].Codecs[0])
	assert.Equal(t, "mp4a.40.2", playlist.Variants[0].Codecs[1])
	assert.Equal(t, 1920, playlist.Variants[0].Resolution.Width)
	assert.Equal(t, 1080, playlist.Variants[0].Resolution.Height)
	assert.Equal(t, "chunked", playlist.Variants[0].Video)
	assert.Equal(t, 1, len(playlist.Variants[0].Alternatives))
	assert.Equal(t, "VIDEO", playlist.Variants[0].Alternatives[0].Type)
	assert.Equal(t, "chunked", playlist.Variants[0].Alternatives[0].GroupID)
	assert.Equal(t, "1080p", playlist.Variants[0].Alternatives[0].Name)
	assert.True(t, playlist.Variants[0].Alternatives[0].Autoselect)
	assert.True(t, playlist.Variants[0].Alternatives[0].Default)

	assert.Equal(t, "http://example.com/720p30/index-dvr.m3u8", playlist.Variants[1].URL)
	assert.Equal(t, 2303475, playlist.Variants[1].Bandwidth)
	assert.Equal(t, 2, len(playlist.Variants[1].Codecs))
	assert.Equal(t, "avc1.4D401F", playlist.Variants[1].Codecs[0])
	assert.Equal(t, "mp4a.40.2", playlist.Variants[1].Codecs[1])
	assert.Equal(t, 1280, playlist.Variants[1].Resolution.Width)
	assert.Equal(t, 720, playlist.Variants[1].Resolution.Height)
	assert.Equal(t, "720p30", playlist.Variants[1].Video)
	assert.Equal(t, 1, len(playlist.Variants[1].Alternatives))
	assert.Equal(t, "VIDEO", playlist.Variants[1].Alternatives[0].Type)
	assert.Equal(t, "720p30", playlist.Variants[1].Alternatives[0].GroupID)
	assert.Equal(t, "720p", playlist.Variants[1].Alternatives[0].Name)
	assert.False(t, playlist.Variants[1].Alternatives[0].Autoselect)
	assert.False(t, playlist.Variants[1].Alternatives[0].Default)
}
