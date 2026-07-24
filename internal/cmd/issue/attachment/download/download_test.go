package download

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputPath(t *testing.T) {
	tempDir := t.TempDir()

	path, err := outputPath("", "report.pdf")
	require.NoError(t, err)
	assert.Equal(t, "report.pdf", path)

	path, err = outputPath(tempDir, "report.pdf")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tempDir, "report.pdf"), path)

	path, err = outputPath("custom.pdf", "report.pdf")
	require.NoError(t, err)
	assert.Equal(t, "custom.pdf", path)

	path, err = outputPath("", "../../report.pdf")
	require.NoError(t, err)
	assert.Equal(t, "report.pdf", path)
}

func TestWriteNewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.txt")

	err := writeNewFile(path, func(dst io.Writer) error {
		_, err := io.WriteString(dst, "attachment contents")
		return err
	})
	require.NoError(t, err)

	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "attachment contents", string(contents))
}

func TestWriteNewFileDoesNotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.txt")
	require.NoError(t, os.WriteFile(path, []byte("original"), 0o644))

	err := writeNewFile(path, func(dst io.Writer) error {
		_, writeErr := io.WriteString(dst, "replacement")
		return writeErr
	})
	require.Error(t, err)

	contents, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, "original", string(contents))
}

func TestWriteNewFileRemovesPartialDownload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.txt")
	wantErr := errors.New("download failed")

	err := writeNewFile(path, func(dst io.Writer) error {
		_, _ = io.WriteString(dst, "partial")
		return wantErr
	})

	assert.ErrorIs(t, err, wantErr)
	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr))
}
