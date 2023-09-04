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

type StreamingProcess struct {
	ctx    context.Context
	cmd    *exec.Cmd
	stdout io.ReadCloser
	buf    *buffer.Buffer
	lock   sync.Mutex
	Done   chan error
}

func NewStreamingProcess(ctx context.Context, bufferSize time.Duration) *StreamingProcess {
	return &StreamingProcess{
		ctx:  ctx,
		buf:  buffer.NewBuffer(bufferSize),
		lock: sync.Mutex{},
		Done: make(chan error, 1),
	}
}

func (p *StreamingProcess) StartProcess() (err error) {
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

	log.Printf("start process: %s", strings.Join(cmdArgs, " "))

	p.cmd = exec.Command(cmdArgs[0], cmdArgs[1:]...)

	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	p.stdout = stdout

	log.Printf("start streaming from %s", url)

	if err = p.cmd.Start(); err != nil {
		return
	}

	log.Printf("started process PID %d, warming up", p.cmd.Process.Pid)

	// wait a little till ffmpeg starts write to the stdout
	time.Sleep(100 * time.Millisecond)

	if err = p.checkProcessState(); err != nil {
		return
	}

	go p.startStatisticsLoop(30 * time.Second)
	go p.startStreamingLoop()

	return
}

func (p *StreamingProcess) checkProcessState() error {
	state := p.cmd.ProcessState

	if state == nil {
		return nil
	}

	if state.Exited() {
		return fmt.Errorf("process exited with code %d", state.ExitCode())
	}

	return nil
}

func (p *StreamingProcess) startStatisticsLoop(interval time.Duration) {
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

func (p *StreamingProcess) startStreamingLoop() {
	for {
		if err := p.checkProcessState(); err != nil {
			p.Done <- err
			return
		}

		chunk := make([]byte, 1024*1024)

		n, err := p.stdout.Read(chunk)
		if n == 0 {
			continue
		}

		if err != nil {
			p.Done <- err
			return
		}

		select {
		case <-p.ctx.Done():
			p.cmd.Process.Signal(os.Interrupt)
			p.Done <- p.ctx.Err()
			return

		default:
		}

		p.lock.Lock()
		p.buf.Trim()
		p.buf.Push(chunk[:n])
		p.lock.Unlock()
	}
}

func (p *StreamingProcess) HandleTimestamp(ts time.Time) (err error) {
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
