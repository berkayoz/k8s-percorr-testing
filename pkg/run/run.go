package run

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Runner executes shell commands, streaming output to the configured writer.
type Runner struct {
	w io.Writer
}

// New returns a Runner that directs all command output to w.
func New(w io.Writer) *Runner {
	return &Runner{w: w}
}

// Cmd runs a command, streaming both stdout and stderr to the writer.
func (r *Runner) Cmd(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = r.w
	cmd.Stderr = r.w
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

// CmdWithDir is like Cmd but sets the working directory.
func (r *Runner) CmdWithDir(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = r.w
	cmd.Stderr = r.w
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

// CmdWithStdin is like Cmd but pipes stdin from the provided byte slice.
func (r *Runner) CmdWithStdin(ctx context.Context, stdin []byte, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = bytes.NewReader(stdin)
	cmd.Stdout = r.w
	cmd.Stderr = r.w
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

// CmdOutput runs a command, captures stdout, and streams stderr to the writer.
func (r *Runner) CmdOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = r.w
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}
	return stdout.Bytes(), nil
}
