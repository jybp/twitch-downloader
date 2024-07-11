package twitch_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jybp/twitch-downloader/twitch"
)

func TestID(t *testing.T) {
	tcs := []struct {
		input        string
		expectedID   string
		expectedType twitch.VideoType
		expectedErr  bool
	}{
		{
			input:        "https://www.twitch.tv/videos/12345",
			expectedID:   "12345",
			expectedType: twitch.TypeVOD,
		},
		{
			input:        "https://www.twitch.tv/videos/12345?some=query&test",
			expectedID:   "12345",
			expectedType: twitch.TypeVOD,
		},
		{
			input:        "https://www.twitch.tv/letsgameitout/video/2182428086",
			expectedID:   "2182428086",
			expectedType: twitch.TypeVOD,
		},
		{
			input:        "https://www.twitch.tv/test/clip/Slug123",
			expectedID:   "Slug123",
			expectedType: twitch.TypeClip,
		},
		{
			input:        "https://clips.twitch.tv/Slug123",
			expectedID:   "Slug123",
			expectedType: twitch.TypeClip,
		},
		{
			input:       "https://www.twitch123.tv/videos/12345",
			expectedErr: true,
		},
		{
			input:       "https://www.twitch123.tv/test",
			expectedErr: true,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.input, func(t *testing.T) {
			id, vType, err := twitch.ID(tc.input)
			if tc.expectedErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedID, id)
			assert.Equal(t, tc.expectedType, vType)
		})
	}
}
