package report

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Dir returns a "reports" subdirectory under the current working directory.
func Dir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, "reports"), nil
}

// WriteToFile creates any parent directories for path, opens the file for
// writing, and delegates the actual content generation to fn.
func WriteToFile(path string, fn func(io.Writer) error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating report directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating report file: %w", err)
	}
	defer f.Close()
	return fn(f)
}
