package stream

import (
	"time"
)

type StreamingProcess interface {
	Start() error
	HandleTimestamp(time.Time) error
	Done() chan error
}
