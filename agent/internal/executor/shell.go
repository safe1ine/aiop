package executor

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

type Shell struct {
	ptmx *os.File
	cmd  *exec.Cmd
}

func NewShell(cols, rows uint16) (*Shell, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}
	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: cols, Rows: rows})
	if err != nil {
		return nil, err
	}
	return &Shell{ptmx: ptmx, cmd: cmd}, nil
}

func (s *Shell) Read(buf []byte) (int, error) {
	return s.ptmx.Read(buf)
}

func (s *Shell) Write(data []byte) (int, error) {
	return s.ptmx.Write(data)
}

func (s *Shell) Resize(cols, rows uint16) error {
	return pty.Setsize(s.ptmx, &pty.Winsize{Cols: cols, Rows: rows})
}

func (s *Shell) Close() {
	s.cmd.Process.Kill()
	s.ptmx.Close()
}
