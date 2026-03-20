package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type noneSandbox struct {
	cfg Config
}

func newNone(cfg Config) Sandbox {
	return &noneSandbox{cfg: cfg}
}

func (n *noneSandbox) Run(ctx context.Context, cmd string, args []string) error {
	fmt.Fprintln(os.Stderr, "WARNING: Running without sandbox isolation. The agent has full filesystem access.")

	c := exec.CommandContext(ctx, cmd, args...)
	c.Dir = n.cfg.WorkDir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	return c.Run()
}
