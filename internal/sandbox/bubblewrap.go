//go:build linux

package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type bubblewrapSandbox struct {
	cfg Config
}

func newBubblewrap(cfg Config) (Sandbox, error) {
	if _, err := exec.LookPath("bwrap"); err != nil {
		return nil, fmt.Errorf("bubblewrap sandbox requires 'bwrap' on PATH: %w", err)
	}
	if cfg.WorkDir == "" {
		return nil, fmt.Errorf("bubblewrap sandbox requires a WorkDir")
	}
	return &bubblewrapSandbox{cfg: cfg}, nil
}

func (b *bubblewrapSandbox) Run(ctx context.Context, cmd string, args []string) error {
	bwrapArgs := []string{
		// Workspace: read-write mount
		"--bind", b.cfg.WorkDir, "/workspace",
		// Read-only system mounts
		"--ro-bind", "/usr", "/usr",
		"--ro-bind", "/bin", "/bin",
		"--ro-bind", "/lib", "/lib",
		"--ro-bind-try", "/lib64", "/lib64",
		// Tmpfs for /tmp
		"--tmpfs", "/tmp",
		// Proc and dev
		"--proc", "/proc",
		"--dev", "/dev",
		// Set working directory
		"--chdir", "/workspace",
		// Unshare all namespaces
		"--unshare-all",
		// Command to run
		cmd,
	}
	bwrapArgs = append(bwrapArgs, args...)

	c := exec.CommandContext(ctx, "bwrap", bwrapArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	// Use process group so we can kill all children on context cancellation
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := c.Start(); err != nil {
		return fmt.Errorf("starting bwrap: %w", err)
	}

	done := make(chan error, 1)
	go func() { done <- c.Wait() }()

	select {
	case <-ctx.Done():
		// Kill the entire process group
		if c.Process != nil {
			_ = syscall.Kill(-c.Process.Pid, syscall.SIGKILL)
		}
		return ctx.Err()
	case err := <-done:
		return err
	}
}
