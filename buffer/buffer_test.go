package buffer_test

import (
	"camrec/buffer"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEmptyBuffer(t *testing.T) {
	b := buffer.NewBuffer(0)
	require.Zero(t, b.Count())
	require.Zero(t, b.Size())
	require.Zero(t, b.Duration())
}

func TestPushItem(t *testing.T) {
	b := buffer.NewBuffer(0)
	b.Push([]byte{0, 1, 2, 3})
	require.Equal(t, 1, b.Count())
	require.Equal(t, 4, b.Size())
	require.InDelta(t, 0, b.Duration(), float64(time.Millisecond))
}

func TestClear(t *testing.T) {
	t.Run("one item", func(t *testing.T) {
		b := buffer.NewBuffer(0)

		b.Push(nil)
		b.Clear()
		require.Zero(t, b.Count())
	})

	t.Run("two items", func(t *testing.T) {
		b := buffer.NewBuffer(0)

		b.Push(nil)
		b.Push(nil)
		b.Clear()
		require.Zero(t, b.Count())
	})
}

func TestTrim(t *testing.T) {
	t.Run("everything", func(t *testing.T) {
		b := buffer.NewBuffer(time.Second)

		now := time.Now()

		b.Put(nil, now.Add(-2*time.Second))
		b.Trim()
		require.Zero(t, b.Count())
	})

	t.Run("partially", func(t *testing.T) {
		b := buffer.NewBuffer(time.Second)

		now := time.Now()

		b.Put(nil, now.Add(-2*time.Second))
		b.Put(nil, now.Add(-100*time.Millisecond))
		b.Put(nil, now)
		b.Trim()
		require.Equal(t, 2, b.Count())
	})
}

func TestNoTrim(t *testing.T) {
	b := buffer.NewBuffer(time.Minute)

	now := time.Now()

	b.Put(nil, now)
	b.Trim()

	require.Equal(t, 1, b.Count())
}

func TestDuration(t *testing.T) {
	t.Run("single item", func(t *testing.T) {
		b := buffer.NewBuffer(time.Second)
		b.Put(nil, time.Now().Add(-100*time.Millisecond))
		require.InDelta(t, 100*time.Millisecond, b.Duration(), float64(time.Millisecond))
	})

	t.Run("with intermediate items", func(t *testing.T) {
		b := buffer.NewBuffer(time.Second)

		b.Put(nil, time.Now().Add(-100*time.Millisecond))
		b.Put(nil, time.Now().Add(-50*time.Millisecond))

		require.InDelta(t, 100*time.Millisecond, b.Duration(), float64(time.Millisecond))
	})
}

func TestUsage(t *testing.T) {
	b := buffer.NewBuffer(time.Minute)

	require.Zero(t, b.Usage())

	b.Put(nil, time.Now().Add(-30*time.Second))
	require.InDelta(t, 50, b.Usage(), 0.01)

	b.Clear()

	b.Put(nil, time.Now().Add(-time.Minute))
	b.Put(nil, time.Now())

	require.Equal(t, float64(100), b.Usage())
}

func TestSearch(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		b := buffer.NewBuffer(0)
		require.Nil(t, b.Search(time.Now()))
	})

	t.Run("lower bound", func(t *testing.T) {
		b := buffer.NewBuffer(time.Second)

		now := time.Now()
		b.Put(nil, now.Add(-time.Second))

		require.Nil(t, b.Search(now.Add(-2*time.Second)))
	})

	t.Run("upper bound ", func(t *testing.T) {
		b := buffer.NewBuffer(time.Second)

		now := time.Now()
		b.Put(nil, now.Add(-time.Second))

		require.Nil(t, b.Search(now))

		b.Put(nil, now)

		require.Nil(t, b.Search(now))
	})

	t.Run("within range", func(t *testing.T) {
		b := buffer.NewBuffer(time.Minute)

		now := time.Now()
		b.Put(nil, now.Add(-30*time.Second))
		b.Put(nil, now.Add(-15*time.Second))

		require.NotNil(t, b.Search(now.Add(-20*time.Second)))
	})

	t.Run("only 2 chunks within range", func(t *testing.T) {
		b := buffer.NewBuffer(time.Minute)

		now := time.Now()
		b.Put([]byte{1}, now.Add(-time.Minute))
		b.Put([]byte{2}, now.Add(-45*time.Second))
		b.Put([]byte{3}, now.Add(-30*time.Second))
		b.Put([]byte{4}, now.Add(-15*time.Second))
		b.Put([]byte{5}, now)

		event := b.Search(now.Add(-40 * time.Second))

		require.NotNil(t, event)
		require.ElementsMatch(t, []byte{2, 3}, event.Data())
	})
}
