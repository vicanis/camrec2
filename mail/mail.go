package mail

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/api/gmail/v1"
)

type Mail struct {
	service       *gmail.Service
	lastMessageId string
	Done          chan error
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

	log.Printf("Gmail service was initialized")

	m = &Mail{
		service: srv,
		Done:    make(chan error, 1),
	}

	return
}

func (m *Mail) StartMessageChecker(ctx context.Context, checkInterval time.Duration) chan time.Time {
	mch := make(chan time.Time)

	go func() {
		ticker := time.NewTicker(checkInterval)

		defer ticker.Stop()
		defer close(mch)

		log.Printf("start mail loop")

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := m.update(mch); err != nil {
					m.Done <- err
					return
				}
			}
		}
	}()

	return mch
}

func (m *Mail) update(mch chan time.Time) (err error) {
	messages, err := m.getUnreadMessages()
	if err != nil {
		return
	}

	for _, msg := range messages {
		tsString := ParseTimestamp(msg.Snippet)
		if tsString == "" {
			continue
		}

		ts, err := BuildTimestamp(tsString)
		if err != nil {
			continue
		}

		mch <- *ts
	}

	return
}

func (m *Mail) getUnreadMessages() ([]*gmail.Message, error) {
	respList, err := m.service.Users.Messages.List("me").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch messages: %w", err)
	}

	messages := make([]*gmail.Message, 0)

	for _, msg := range respList.Messages {
		if m.lastMessageId == msg.Id {
			break
		}

		respMsg, err := m.service.Users.Messages.Get("me", msg.Id).Do()
		if err != nil {
			return nil, err
		}

		if !isHicloudMessage(respMsg) {
			continue
		}

		messages = append(messages, respMsg)
	}

	for _, msg := range respList.Messages {
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

func isHicloudMessage(msg *gmail.Message) bool {
	if msg.Payload == nil || msg.Payload.Headers == nil {
		return false
	}

	for _, v := range msg.Payload.Headers {
		if v.Name == "From" && v.Value == "no_reply@hicloudcam.com" {
			return true
		}
	}

	return false
}
