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
		c.timestamp.Format("02.01.2006 15:04:05"),
		len(c.data),
	)
}
