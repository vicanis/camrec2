package buffer

import (
	"fmt"
	"time"
)

type chunk struct {
	offset    int
	length    int
	timestamp time.Time
}

func (c chunk) String() string {
	return fmt.Sprintf(
		"%s: %d bytes",
		c.timestamp.Format(time.RFC1123),
		c.length,
	)
}
