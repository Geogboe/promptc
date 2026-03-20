package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNew_InvalidType(t *testing.T) {
	_, err := New(Config{Type: "invalid"})
	if err == nil {
		t.Fatal("expected error for unknown sandbox type")
	}
}

func TestNew_EmptyType(t *testing.T) {
	// Empty type defaults to none
	s, err := New(Config{Type: ""})
	if err != nil {
		t.Fatalf("unexpected error for empty type: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sandbox for empty type")
	}
}

func TestNew_None(t *testing.T) {
	s, err := New(Config{Type: "none"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sandbox")
	}
}

func TestNone_RunsCommand(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(Config{Type: "none", WorkDir: dir})

	// Write a marker file via the "none" sandbox
	outFile := filepath.Join(dir, "marker.txt")
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "hello", ">", outFile}
	} else {
		cmd = "sh"
		args = []string{"-c", "echo hello > " + outFile}
	}

	if err := s.Run(context.Background(), cmd, args); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		t.Error("expected marker file to be created")
	}
}

func TestNone_ContextCancellation(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(Config{Type: "none", WorkDir: dir})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "timeout", "/t", "30"}
	} else {
		cmd = "sleep"
		args = []string{"30"}
	}

	err := s.Run(ctx, cmd, args)
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
}

func TestBubblewrap_UnavailableOnNonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("skipping non-Linux bubblewrap stub test on Linux")
	}
	_, err := New(Config{Type: "bubblewrap"})
	if err == nil {
		t.Fatal("expected error for bubblewrap on non-Linux")
	}
}

func TestDocker_RequiresImage(t *testing.T) {
	_, err := New(Config{Type: "docker", Image: ""})
	if err == nil {
		t.Fatal("expected error for docker without image")
	}
}

func TestDocker_RequiresWorkDir(t *testing.T) {
	_, err := New(Config{Type: "docker", Image: "ubuntu:22.04", WorkDir: ""})
	if err == nil {
		t.Fatal("expected error for docker without WorkDir")
	}
}
