package mail

import (
	"fmt"
	"time"

	"google.golang.org/api/gmail/v1"
)

type Mail struct {
	service       *gmail.Service
	lastMessageId string
}

type Message struct {
	Id        string
	Timestamp time.Time
}

func Initialize() (m *Mail, err error) {
	srv, err := GetService()
	if err != nil {
		return
	}

	m = &Mail{service: srv}

	return
}

func (m *Mail) GetHicloudMessages() ([]Message, error) {
	messages, err := m.getMessages()
	if err != nil {
		return nil, err
	}

	list := make([]Message, 0)

	for _, msg := range messages {
		isHicloud := false

		for _, v := range msg.Payload.Headers {
			if v.Name == "From" && v.Value == "no_reply@hicloudcam.com" {
				isHicloud = true
				break
			}
		}

		if !isHicloud {
			continue
		}

		str := ParseTimestamp(msg.Snippet)
		if str == "" {
			continue
		}

		timestamp := BuildTimestamp(str)
		if timestamp == nil {
			continue
		}

		list = append(list, Message{
			Id:        msg.Id,
			Timestamp: *timestamp,
		})
	}

	return list, nil
}

func (m *Mail) getMessages() ([]*gmail.Message, error) {
	r, err := m.service.Users.Messages.List("me").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch messages: %s", err)
	}

	messages := make([]*gmail.Message, 0)

	for _, msg := range r.Messages {
		if m.lastMessageId == msg.Id {
			break
		}

		r2, err := m.service.Users.Messages.Get("me", msg.Id).Do()
		if err != nil {
			return nil, err
		} else {
			messages = append(messages, r2)
		}
	}

	for _, msg := range r.Messages {
		if m.lastMessageId == msg.Id {
			break
		}

		if m.lastMessageId == "" || m.lastMessageId != msg.Id {
			m.lastMessageId = msg.Id
			break
		}
	}

	return messages, nil
}
