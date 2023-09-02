package mail

import (
	"regexp"
	"time"
)

func ParseTimestamp(text string) string {
	rx := regexp.MustCompile(`(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})`)

	list := rx.FindAll([]byte(text), -1)

	if len(list) == 0 {
		return ""
	}

	str := string(list[0])

	return str
}

func BuildTimestamp(text string) (ts *time.Time, err error) {
	t, err := time.ParseInLocation(time.DateTime, text, time.Local)
	if err != nil {
		return
	}

	ts = &t

	return
}
