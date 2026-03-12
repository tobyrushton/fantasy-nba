package scraper_test

import (
	"io"
	"os"
	"testing"
)

func getReadCloser(t *testing.T, filePath string) io.ReadCloser {
	t.Helper()
	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open test file %s: %v", filePath, err)
	}
	return io.NopCloser(f)
}
