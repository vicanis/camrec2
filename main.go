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

	m, err := mail.Initialize()
	if err != nil {
		log.Printf("mail initialize failed: %s", err)
		cancel()
		return
	}

	tschan := m.StartMessageChecker(ctx, 5*time.Second)

	go func() {
		p := stream.NewFfmpegStreamer(ctx, 120*time.Second)

		if err := p.Start(); err != nil {
			log.Printf("streaming start failed: %s", err)
			cancel()
			return
		}

		for {
			select {
			case <-ctx.Done():
				return

			case ts := <-tschan:
				go func(ts time.Time) {
					time.Sleep(20 * time.Second)

					log.Printf("handle timestamp: %s", ts.Format(time.RFC1123))

					if err := p.HandleTimestamp(ts); err != nil {
						log.Printf("> failed: %s", err)
					}
				}(ts)

			case err := <-p.Done():
				log.Printf("streaming end: %s", err)
				cancel()
				return

			case err := <-m.Done:
				log.Printf("message loop end: %s", err)
				cancel()
				return
			}
		}
	}()

	time.Sleep(time.Second)

	log.Printf("press ctrl+c to interrupt")

	go func() {
		sig := <-sigchan
		log.Printf("signal: %s", sig)
		cancel()
	}()

	<-ctx.Done()

	time.Sleep(time.Second)
}
