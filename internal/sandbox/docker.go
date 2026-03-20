package sandbox

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type dockerSandbox struct {
	cfg Config
}

func newDocker(cfg Config) (Sandbox, error) {
	if cfg.Image == "" {
		return nil, fmt.Errorf("docker sandbox requires an image")
	}
	if cfg.WorkDir == "" {
		return nil, fmt.Errorf("docker sandbox requires a WorkDir")
	}
	return &dockerSandbox{cfg: cfg}, nil
}

func (d *dockerSandbox) Run(ctx context.Context, cmd string, args []string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("creating docker client: %w\n\nEnsure Docker Engine is running.\n  Windows: winget install Docker.DockerEngine  (or choco install docker-engine)\n  macOS:   brew install docker\n  Linux:   your package manager (apt/dnf/pacman)", err)
	}
	defer func() { _ = cli.Close() }()

	// Pull image (no-op if already present)
	fmt.Fprintf(os.Stderr, "· Pulling image %s...\n", d.cfg.Image) //nolint:errcheck // stderr progress
	reader, err := cli.ImagePull(ctx, d.cfg.Image, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("pulling image %s: %w", d.cfg.Image, err)
	}
	_, _ = io.Copy(io.Discard, reader) // drain pull progress stream
	_ = reader.Close()

	// Build command slice
	fullCmd := append([]string{cmd}, args...)

	// Create container
	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image:      d.cfg.Image,
			Cmd:        fullCmd,
			WorkingDir: "/workspace",
		},
		&container.HostConfig{
			Binds: []string{d.cfg.WorkDir + ":/workspace"},
		},
		nil, nil, "",
	)
	if err != nil {
		return fmt.Errorf("creating container: %w", err)
	}

	containerID := resp.ID

	// Ensure cleanup on exit or context cancellation
	defer func() {
		stopCtx := context.Background()
		timeout := 10
		_ = cli.ContainerStop(stopCtx, containerID, container.StopOptions{Timeout: &timeout})
		_ = cli.ContainerRemove(stopCtx, containerID, container.RemoveOptions{Force: true})
	}()

	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("starting container: %w", err)
	}

	// Stream logs
	logOpts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	}
	logs, err := cli.ContainerLogs(ctx, containerID, logOpts)
	if err != nil {
		return fmt.Errorf("attaching to container logs: %w", err)
	}
	defer func() { _ = logs.Close() }()
	_, _ = stdcopy.StdCopy(os.Stdout, os.Stderr, logs)

	// Wait for completion
	statusCh, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("waiting for container: %w", err)
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("container exited with code %d", status.StatusCode)
		}
	}

	return nil
}
