package mail

import (
	"log"
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

func BuildTimestamp(text string) *time.Time {
	t, err := time.ParseInLocation(time.DateTime, text, time.Local)
	if err != nil {
		log.Printf("parse failed: %s", err)
		return nil
	}

	return &t
}
