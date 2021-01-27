package twitch_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jybp/twitch-downloader/twitch"
)

func TestID(t *testing.T) {
	id, err := twitch.ID("https://www.twitch.tv/videos/12345")
	assert.Nil(t, err)
	assert.Equal(t, "12345", id)
}

func TestID_Query(t *testing.T) {
	id, err := twitch.ID("https://www.twitch.tv/videos/12345?some=query&test")
	assert.Nil(t, err)
	assert.Equal(t, "12345", id)
}

func TestID_WrongHost(t *testing.T) {
	_, err := twitch.ID("https://www.twitch123.tv/videos/12345")
	assert.NotNil(t, err)
}

func TestID_Clip(t *testing.T) {
	_, err := twitch.ID("https://www.twitch.tv/test/clip/12345")
	assert.NotNil(t, err)
}

func TestID_Stream(t *testing.T) {
	_, err := twitch.ID("https://www.twitch.tv/test")
	assert.NotNil(t, err)
}
