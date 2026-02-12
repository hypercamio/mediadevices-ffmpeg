package ffmpeg

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

const stderrBufSize = 4096

// ffmpegProcess manages a running FFmpeg subprocess.
type ffmpegProcess struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
	cancel context.CancelFunc

	stderrMu  sync.Mutex
	stderrBuf []byte
	done      chan struct{}
}

// startProcess launches an FFmpeg subprocess with the given arguments.
// Stdout is available for reading via Read(). Stderr is drained into a
// circular buffer accessible via LastStderr().
func startProcess(ffmpegPath string, args []string) (*ffmpegProcess, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, ffmpegPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("ffmpeg stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("ffmpeg stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("ffmpeg start: %w", err)
	}

	p := &ffmpegProcess{
		cmd:    cmd,
		stdout: stdout,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	// Drain stderr in background, keeping the last stderrBufSize bytes.
	go p.drainStderr(stderr)

	return p, nil
}

func (p *ffmpegProcess) drainStderr(r io.Reader) {
	defer close(p.done)
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			p.stderrMu.Lock()
			p.stderrBuf = append(p.stderrBuf, buf[:n]...)
			if len(p.stderrBuf) > stderrBufSize {
				p.stderrBuf = p.stderrBuf[len(p.stderrBuf)-stderrBufSize:]
			}
			p.stderrMu.Unlock()
		}
		if err != nil {
			return
		}
	}
}

// Read reads from the FFmpeg subprocess stdout.
func (p *ffmpegProcess) Read(buf []byte) (int, error) {
	return p.stdout.Read(buf)
}

// Stop terminates the FFmpeg subprocess.
func (p *ffmpegProcess) Stop() error {
	p.cancel()
	// Wait for stderr drain to finish so we capture final output.
	<-p.done
	return p.cmd.Wait()
}

// LastStderr returns the last portion of FFmpeg's stderr output,
// useful for diagnosing errors.
func (p *ffmpegProcess) LastStderr() string {
	p.stderrMu.Lock()
	defer p.stderrMu.Unlock()
	return string(p.stderrBuf)
}
