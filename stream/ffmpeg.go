package stream

import (
	"camrec/buffer"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type FfmpegStreamer struct {
	ctx    context.Context
	cmd    *exec.Cmd
	stdout io.ReadCloser
	buf    *buffer.Buffer
	lock   sync.Mutex
	done   chan error
}

func NewFfmpegStreamer(ctx context.Context, bufferSize time.Duration) StreamingProcess {
	return &FfmpegStreamer{
		ctx:  ctx,
		buf:  buffer.NewBuffer(bufferSize),
		lock: sync.Mutex{},
		done: make(chan error, 1),
	}
}

func (p *FfmpegStreamer) Start() (err error) {
	url := os.Getenv("STREAM")

	if url == "" {
		err = errors.New("no stream URL")
		return
	}

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

	log.Printf("start streamer process: %s", strings.Join(cmdArgs, " "))

	p.cmd = exec.Command(cmdArgs[0], cmdArgs[1:]...)

	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	p.stdout = stdout

	if err = p.cmd.Start(); err != nil {
		return
	}

	log.Printf("streamer process was started: PID %d", p.cmd.Process.Pid)

	// wait a little till ffmpeg starts write to the stdout
	time.Sleep(100 * time.Millisecond)

	if err = p.checkProcessState(); err != nil {
		return
	}

	go p.startStatisticsLoop(30 * time.Second)
	go p.startStreamingLoop()

	return
}

func (p *FfmpegStreamer) HandleTimestamp(ts time.Time) (err error) {
	p.lock.Lock()
	event := p.buf.Search(ts)
	p.lock.Unlock()

	if event != nil {
		if err = event.SaveFile(); err != nil {
			err = fmt.Errorf("event file save failed: %w", err)
			return
		}
	}

	return
}

func (p *FfmpegStreamer) Done() chan error {
	return p.done
}

func (p *FfmpegStreamer) startStatisticsLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			log.Printf(
				"chunk count %d, size %d, duration %f sec (usage %.2f%%)",
				p.buf.Count(), p.buf.Size(),
				p.buf.Duration().Seconds(), p.buf.Usage(),
			)
		}
	}
}

func (p *FfmpegStreamer) startStreamingLoop() {
	for {
		if err := p.checkProcessState(); err != nil {
			p.done <- err
			return
		}

		chunk := make([]byte, 1024*1024)

		n, err := p.stdout.Read(chunk)
		if n == 0 {
			continue
		}

		if err != nil {
			p.done <- err
			return
		}

		select {
		case <-p.ctx.Done():
			p.cmd.Process.Signal(os.Interrupt)
			p.done <- p.ctx.Err()
			return

		default:
		}

		p.lock.Lock()
		p.buf.Trim()
		p.buf.Push(chunk[:n])
		p.lock.Unlock()
	}
}

func (p *FfmpegStreamer) checkProcessState() error {
	state := p.cmd.ProcessState

	if state == nil {
		return nil
	}

	if state.Exited() {
		return fmt.Errorf("process exited with code %d", state.ExitCode())
	}

	return nil
}
