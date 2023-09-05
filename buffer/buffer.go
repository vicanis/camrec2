package buffer

import (
	"camrec/event"
	"math"
	"time"
)

type Buffer struct {
	data     []byte
	chunks   []chunk
	duration time.Duration
}

func NewBuffer(duration time.Duration) *Buffer {
	return &Buffer{
		data:     make([]byte, 0),
		chunks:   make([]chunk, 0),
		duration: duration,
	}
}

func (b *Buffer) Put(data []byte, ts time.Time) {
	b.chunks = append(b.chunks, chunk{
		offset:    len(b.data),
		length:    len(data),
		timestamp: ts,
	})

	b.data = append(b.data, data...)
}

func (b *Buffer) Push(data []byte) {
	b.Put(data, time.Now())
}

func (b *Buffer) Trim() {
	trimStart := -1

	lbound := time.Now().Add(-b.duration)

	offsetShift := 0

	for i, chunk := range b.chunks {
		if chunk.timestamp.Before(lbound) {
			trimStart = i
			offsetShift += chunk.length
		}
	}

	if trimStart >= 0 {
		for i := range b.chunks {
			if i > trimStart {
				b.chunks[i].offset -= offsetShift
			}
		}

		b.chunks = b.chunks[trimStart+1:]
		b.data = b.data[offsetShift:]
	}
}

func (b *Buffer) Clear() {
	b.chunks = make([]chunk, 0)
}

func (b Buffer) Count() int {
	return len(b.chunks)
}

func (b Buffer) Size() (size int) {
	for _, chunk := range b.chunks {
		size += chunk.length
	}

	return
}

func (b Buffer) Duration() (dur time.Duration) {
	for i, chunk := range b.chunks {
		if i < len(b.chunks)-1 {
			dur += b.chunks[i+1].timestamp.Sub(chunk.timestamp)
		} else {
			dur += time.Since(chunk.timestamp)
		}
	}

	return
}

func (b Buffer) Usage() float64 {
	return math.Min(100, 100*float64(b.Duration())/float64(b.duration))
}

// Search chunks before and after ts
func (b Buffer) Search(ts time.Time) *event.Event {
	if len(b.chunks) == 0 {
		return nil
	}

	// if ts is not in range
	if b.chunks[0].timestamp.After(ts) || b.chunks[len(b.chunks)-1].timestamp.Before(ts) {
		return nil
	}

	lboundTime := ts.Add(-30 * time.Second)
	uboundTime := ts.Add(30 * time.Second)

	lboundIndex := 0
	uboundIndex := len(b.chunks) - 1

	for i, chunk := range b.chunks {
		if chunk.timestamp.After(lboundTime) {
			lboundIndex = i
			break
		}
	}

	for j := len(b.chunks) - 1; j > lboundIndex; j-- {
		if b.chunks[j].timestamp.Before(uboundTime) {
			uboundIndex = j
			break
		}
	}

	offsetStart := b.chunks[lboundIndex].offset

	offsetEnd := offsetStart
	for i := lboundIndex; i <= uboundIndex; i++ {
		offsetEnd += b.chunks[i].length
	}

	chunkPart := b.data[offsetStart:offsetEnd]

	found := make([]byte, len(chunkPart))
	copy(found, chunkPart)

	return event.NewEvent(ts, found)

}
