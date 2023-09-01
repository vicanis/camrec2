package stream

import (
	"camrec/buffer"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func Start(ctx context.Context, tschan chan time.Time) error {
	url := os.Getenv("STREAM")

	cmdArgs := []string{
		"ffmpeg",
		"-i",
		url,
		"-v",
		"0",
		"-f",
		"h264",
		"-c",
		"copy",
		"-",
	}

	log.Printf("start process: %s", strings.Join(cmdArgs, " "))

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

	streamReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	defer streamReader.Close()

	log.Printf("start streaming from %s", url)

	if err := cmd.Start(); err != nil {
		return err
	}

	defer func() {
		if cmd.Process != nil {
			cmd.Process.Release()
		}
	}()

	log.Printf("started process PID %d, warming up", cmd.Process.Pid)

	// wait a little till ffmpeg starts write to the stdout
	time.Sleep(100 * time.Millisecond)

	if err := checkProcessState(cmd); err != nil {
		return err
	}

	buffer := buffer.NewBuffer(120 * time.Second)

	go func() {
		ticker := time.NewTicker(30 * time.Second)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Printf(
					"records: %d, size %d, duration %f sec (usage %.2f%%)",
					buffer.Count(), buffer.Size(),
					buffer.Duration().Seconds(), buffer.Usage(),
				)
			}
		}
	}()

loop:
	for {
		if err := checkProcessState(cmd); err != nil {
			return err
		}

		buf := make([]byte, 1024*1024)

		n, err := streamReader.Read(buf)
		if n == 0 {
			continue
		}

		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			cmd.Process.Signal(os.Interrupt)
			break loop

		case ts := <-tschan:
			log.Printf("streamer: process timestamp %s", ts.Format("15:04:05 02-01-2006"))

			event := buffer.Search(ts)

			if event != nil {
				fname := "events/" + ts.Format("15:04:05 02-01-2006")

				f, err := os.OpenFile(fname, os.O_CREATE, 0644)
				if err != nil {
					log.Printf("event file create failed: %s", err)
					break
				}

				n, err := f.Write(event)
				if err != nil {
					log.Printf("event file write failed: %s", err)
					break
				}

				log.Printf("event file was created: %s (%d bytes)", fname, n)
			} else {
				log.Printf("no event data")
			}

		default:
			// pass
		}

		buffer.Trim()
		buffer.Push(buf[:n])
	}

	log.Printf("streaming end: %s", ctx.Err())

	return nil
}

func checkProcessState(cmd *exec.Cmd) error {
	p := cmd.ProcessState

	if p == nil {
		return nil
	}

	if p.Exited() {
		return fmt.Errorf("process exited with code %d", p.ExitCode())
	}

	return nil
}
