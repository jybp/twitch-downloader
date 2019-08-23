package m3u8_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jybp/twitch-downloader/m3u8"
)

func TestMedia(t *testing.T) {
	b := []byte(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:15
#EXT-X-PLAYLIST-TYPE:EVENT
#EXT-X-MEDIA-SEQUENCE:2
#EXT-X-TWITCH-ELAPSED-SECS:0.000
#EXT-X-TWITCH-TOTAL-SECS:578.690
#EXTINF:11.5,
0.ts?query=val
#EXTINF:13
http://custom.com/720p30/1.ts?query=val
#EXT-X-ENDLIST`)
	playlist, err := m3u8.Media(bytes.NewReader(b), "http://example.com/720p30/index-dvr.m3u8")
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if testing.Verbose() {
		t.Logf("%+v", playlist)
	}
	assert.Equal(t, time.Second*15, playlist.TargetDuration)
	assert.Equal(t, "EVENT", playlist.Type)
	assert.Equal(t, 2, playlist.Sequence)
	assert.True(t, playlist.Ended)

	assert.Equal(t, 2, len(playlist.Segments))

	assert.Equal(t, time.Second*11+time.Millisecond*500, playlist.Segments[0].Duration)
	assert.Equal(t, 2, playlist.Segments[0].Number)
	assert.Equal(t, "http://example.com/720p30/0.ts?query=val", playlist.Segments[0].URL)

	assert.Equal(t, time.Second*13, playlist.Segments[1].Duration)
	assert.Equal(t, 3, playlist.Segments[1].Number)
	assert.Equal(t, "http://custom.com/720p30/1.ts?query=val", playlist.Segments[1].URL)
}
