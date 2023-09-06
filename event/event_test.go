package event_test

import (
	"camrec/event"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	e := event.NewEvent(time.Now(), nil)
	require.NotNil(t, e)
}

func TestData(t *testing.T) {
	e := event.NewEvent(time.Now(), []byte{1, 2, 3})
	require.EqualValues(t, []byte{1, 2, 3}, e.Data())
}

func TestSaveFile(t *testing.T) {
	now := time.Now()

	t.Run("no data", func(t *testing.T) {
		e := event.NewEvent(now, nil)

		event.OutputDirectory = t.TempDir()

		err := e.SaveFile()

		require.Error(t, err)
	})

	t.Run("save file", func(t *testing.T) {
		filesCount := func() []string {
			dir := event.Directory()

			entries, err := os.ReadDir(dir)

			if err != nil {
				t.Fatal(err)
			}

			files := make([]string, 0)

			for _, entry := range entries {
				files = append(files, dir+"/"+entry.Name())
			}

			return files
		}

		event.OutputDirectory = t.TempDir()

		e := event.NewEvent(now, []byte{1, 2, 3})

		err := e.SaveFile()
		require.NoError(t, err)
		require.Equal(t, 1, len(filesCount()))

		err = e.SaveFile()
		require.NoError(t, err)
		require.Equal(t, 2, len(filesCount()))

		err = e.SaveFile()
		require.NoError(t, err)
		require.Equal(t, 3, len(filesCount()))
	})

	t.Run("events directory is not writable", func(t *testing.T) {
		event.OutputDirectory = t.TempDir()

		dir := event.Directory()

		err := os.Mkdir(dir, 0000)
		require.NoError(t, err)

		e := event.NewEvent(time.Now(), []byte{1, 2, 3})
		err = e.SaveFile()
		require.Error(t, err)
	})

	t.Run("events directory create failure", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := tmpDir + "/subdir"

		err := os.Mkdir(subDir, 0000)
		require.NoError(t, err)

		err = os.Chmod(subDir, 0000)
		require.NoError(t, err)

		event.OutputDirectory = subDir

		e := event.NewEvent(time.Now(), []byte{1, 2, 3})
		err = e.SaveFile()
		require.Error(t, err)
	})

	t.Run("save blank file", func(t *testing.T) {
		event.OutputDirectory = t.TempDir()

		e := event.NewEvent(time.Now(), []byte{})
		err := e.SaveFile()
		require.Error(t, err)
	})
}
