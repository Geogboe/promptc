//go:build !linux

package sandbox

import "fmt"

// newBubblewrap is a stub for non-Linux platforms.
func newBubblewrap(cfg Config) (Sandbox, error) {
	return nil, fmt.Errorf("bubblewrap sandbox is only available on Linux; use docker or none instead")
}
