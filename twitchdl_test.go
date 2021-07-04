package twitchdl

import (
	"fmt"
	"testing"
	"time"

	"github.com/jybp/twitch-downloader/m3u8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSliceSegments(t *testing.T) {
	segments := []m3u8.MediaSegment{
		{Number: 0, Duration: time.Second * 5},
		{Number: 1, Duration: time.Second * 5},
		{Number: 2, Duration: time.Second * 5},
		{Number: 3, Duration: time.Second * 5},
	}

	tcs := []struct {
		start       time.Duration
		end         time.Duration
		expected    []int
		expectedErr bool
	}{
		{
			start:       time.Second * 2,
			end:         time.Second,
			expectedErr: true,
		},
		{
			start:       time.Second * 8,
			end:         time.Second * 8,
			expectedErr: true,
		},
		{
			start:       -time.Second * 2,
			end:         -time.Second,
			expectedErr: true,
		},
		{
			start:       time.Second * 20,
			end:         time.Second * 21,
			expectedErr: true,
		},
		{
			start:    time.Duration(0),
			end:      time.Duration(0),
			expected: []int{0, 1, 2, 3},
		},
		{
			start:    time.Second * 5,
			end:      time.Duration(0),
			expected: []int{1, 2, 3},
		},
		{
			start:    time.Duration(0),
			end:      time.Second * 5,
			expected: []int{0},
		},
		{
			start:    time.Second * 6,
			end:      time.Second * 10,
			expected: []int{1},
		},
		{
			start:    time.Second * 6,
			end:      time.Second * 11,
			expected: []int{1, 2},
		},
		{
			start:    time.Second * 6,
			end:      time.Second * 22,
			expected: []int{1, 2, 3},
		},
	}

	for _, tc := range tcs {
		t.Run(fmt.Sprintf("start: %v end: %v", tc.start, tc.end), func(t *testing.T) {
			actual, err := sliceSegments(segments, tc.start, tc.end)
			if tc.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, actual, len(tc.expected))
			for i := 0; i < len(actual); i++ {
				assert.Equal(t, tc.expected[i], actual[i].Number)
			}
		})
	}
}
