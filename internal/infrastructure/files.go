package infrastructure

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func writeFile(path string, write func(io.Writer) error) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".pos-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	writeErr := write(file)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return fmt.Errorf("сохранение %q: %w", path, err)
	}
	return nil
}
