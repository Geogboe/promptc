// Package sandbox provides execution isolation for the build agent.
package sandbox

import (
	"context"
	"fmt"
)

// Config describes how to construct a sandbox.
type Config struct {
	Type    string // docker | bubblewrap | none
	Image   string // docker image (required for docker type)
	WorkDir string // host directory to mount as /workspace
}

// Sandbox executes a command with filesystem isolation.
type Sandbox interface {
	// Run executes cmd with args inside the sandbox.
	// WorkDir is mounted as the working directory.
	// stdout/stderr from the command are streamed to the terminal.
	Run(ctx context.Context, cmd string, args []string) error
}

// New returns the Sandbox implementation for config.Type.
func New(cfg Config) (Sandbox, error) {
	switch cfg.Type {
	case "docker":
		return newDocker(cfg)
	case "bubblewrap":
		return newBubblewrap(cfg)
	case "none":
		return newNone(cfg), nil
	case "":
		return newNone(cfg), nil
	default:
		return nil, fmt.Errorf("unknown sandbox type %q: must be one of docker, bubblewrap, none", cfg.Type)
	}
}
