package event

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type Event struct {
	ts   time.Time
	data []byte
}

var OutputDirectory = "."

func NewEvent(ts time.Time, data []byte) *Event {
	return &Event{
		ts:   ts,
		data: data,
	}
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func Directory() string {
	return OutputDirectory + "/events"
}

func (e Event) FileName() string {
	base := Directory() + "/" + e.ts.Format("2006-01-02_15-04-05")

	index := 0
	for {
		if index == 0 {
			if !isExist(base) {
				break
			}

			index++
		}

		check := fmt.Sprintf("%s-%d", base, index)

		if !isExist(check) {
			return check
		}

		index++
	}

	return base
}

func (e Event) SaveFile() (err error) {
	if e.data == nil || len(e.data) == 0 {
		return errors.New("empty event data")
	}

	dir := Directory()

	if !isExist(dir) {
		if err = os.Mkdir(dir, 0777); err != nil {
			return
		}

		err = nil
	}

	f, err := os.OpenFile(e.FileName(), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}

	defer f.Close()

	_, err = f.Write(e.data)
	if err != nil {
		return
	}

	return nil
}

func (e Event) Data() []byte {
	return e.data
}
