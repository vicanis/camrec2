package buffer

import (
	"fmt"
	"time"
)

type chunk struct {
	data      []byte
	timestamp time.Time
}

func (c chunk) String() string {
	return fmt.Sprintf(
		"%s: %d bytes",
		c.timestamp.Format(time.RFC1123),
		len(c.data),
	)
}
