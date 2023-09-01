package buffer

import "time"

type chunk struct {
	data      []byte
	timestamp time.Time
}
