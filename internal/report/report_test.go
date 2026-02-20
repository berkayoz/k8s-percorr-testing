package report

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteToFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "sub", "report.txt")

	err := WriteToFile(outPath, func(w io.Writer) error {
		_, err := w.Write([]byte("hello"))
		return err
	})
	if err != nil {
		t.Fatalf("WriteToFile: %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("expected %q, got %q", "hello", string(got))
	}
}
