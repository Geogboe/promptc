// Package progress provides simple CLI progress output for long-running operations.
// Output goes to stderr so stdout remains clean for machine-readable content.
package progress

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Writer is the destination for progress output (defaults to os.Stderr).
var Writer io.Writer = os.Stderr

// Step prints a progress step message with a leading "·" indicator.
func Step(format string, args ...any) {
	fmt.Fprintf(Writer, "· "+format+"\n", args...)
}

// Done prints a completion message with a "✓" indicator.
func Done(format string, args ...any) {
	fmt.Fprintf(Writer, "✓ "+format+"\n", args...)
}

// Fail prints a failure message with a "✗" indicator.
func Fail(format string, args ...any) {
	fmt.Fprintf(Writer, "✗ "+format+"\n", args...)
}

// Spinner runs a simple inline spinner while fn executes.
// It prints msg and shows elapsed time on completion or failure.
func Spinner(msg string, fn func() error) error {
	fmt.Fprintf(Writer, "· %s... ", msg)
	start := time.Now()

	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	done := make(chan error, 1)

	go func() {
		done <- fn()
	}()

	i := 0
	for {
		select {
		case err := <-done:
			elapsed := time.Since(start).Round(time.Millisecond)
			if err != nil {
				fmt.Fprintf(Writer, "\r✗ %s (%s)\n", msg, elapsed)
				return err
			}
			fmt.Fprintf(Writer, "\r✓ %s (%s)\n", msg, elapsed)
			return nil
		default:
			fmt.Fprintf(Writer, "\r%s %s... ", frames[i%len(frames)], msg)
			time.Sleep(80 * time.Millisecond)
			i++
		}
	}
}
