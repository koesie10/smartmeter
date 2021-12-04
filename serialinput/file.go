package serialinput

import (
	"fmt"
	"io"
	"os"
)

type FileOptions struct {
	Filename string `env:"FILENAME" flag:"filename" desc:"filename to read from"`
}

func OpenFile(opts *FileOptions) (io.ReadCloser, error) {
	file, err := os.Open(opts.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", opts.Filename, err)
	}

	return file, nil
}
