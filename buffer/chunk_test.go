package buffer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFormatChunk(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		c := chunk{}
		require.Equal(
			t,
			"Mon, 01 Jan 0001 00:00:00 UTC: 0 bytes",
			c.String(),
		)
	})

	t.Run("not empty", func(t *testing.T) {
		ts, err := time.Parse(time.RFC3339, "2012-03-10T14:05:22+04:00")

		require.NoError(t, err)

		c := chunk{
			data:      []byte{1, 2, 3, 4},
			timestamp: ts,
		}

		require.Equal(t, "Sat, 10 Mar 2012 14:05:22 +04: 4 bytes", c.String())
	})
}
