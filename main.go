package main

import (
	"camrec/mail"
	"camrec/stream"
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	url := os.Getenv("STREAM")

	if url == "" {
		log.Fatal("no stream URL")
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)

	tschan := make(chan time.Time)

	go func() {
		err := stream.Start(ctx, tschan)
		if err != nil {
			log.Printf("streaming failed: %s", err)
			cancel()
		}
	}()

	go func() {
		m, err := mail.Initialize()
		if err != nil {
			log.Printf("mail initialize failed: %s", err)
			cancel()
			return
		}

		log.Printf("start mail loop")

		ticker := time.Tick(5 * time.Second)

	outer:
		for {
			select {
			case <-ctx.Done():
				break outer
			case <-ticker:
				// pass
			}

			messages, err := m.GetHicloudMessages()
			if err != nil {
				log.Printf("get message failed: %s", err)
				continue
			}

			for _, msg := range messages {
				tschan <- msg.Timestamp
			}
		}

		close(tschan)

		log.Printf("end mail loop")
	}()

	log.Printf("press ctrl+c to interrupt")

loop:
	for {
		select {
		case <-sigchan:
			cancel()
			break loop
		case <-ctx.Done():
			break loop
		}
	}

	time.Sleep(time.Second)
}
