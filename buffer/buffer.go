package buffer

import "time"

type Buffer struct {
	chunks []chunk
	limit  time.Duration
}

func NewBuffer(limit time.Duration) Buffer {
	return Buffer{
		chunks: make([]chunk, 0),
		limit:  limit,
	}
}

func (b *Buffer) Push(data []byte) {
	b.chunks = append(b.chunks, chunk{
		data:      data,
		timestamp: time.Now(),
	})
}

func (b *Buffer) Trim() {
	trimStart := -1

	for i, chunk := range b.chunks {
		if chunk.timestamp.Before(time.Now().Add(-b.limit)) {
			trimStart = i
		}
	}

	if trimStart >= 0 {
		b.chunks = b.chunks[trimStart:]
	}
}

func (b Buffer) Count() int {
	return len(b.chunks)
}

func (b Buffer) Size() (size int) {
	for _, chunk := range b.chunks {
		size += len(chunk.data)
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
	return 100 * float64(b.Duration()) / float64(b.limit)
}

// Search chunks before and after ts
func (b Buffer) Search(ts time.Time) []byte {
	if len(b.chunks) == 0 {
		return nil
	}

	// if ts is not in range
	if b.chunks[0].timestamp.After(ts) || b.chunks[len(b.chunks)-1].timestamp.Before(ts) {
		return nil
	}

	// b.chunks stores chunks in ascending order
	// (m:s)    03:00   04:00   05:00
	// search for   03:30
	for i, chunk := range b.chunks {
		if chunk.timestamp.After(ts) {
			if i == 0 {
				return nil
			}

			found := make([]byte, 0)

			found = append(found, b.chunks[i-1].data...)
			found = append(found, chunk.data...)

			return found
		}
	}

	return nil
}
