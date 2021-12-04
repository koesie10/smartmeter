package serialinput

import (
	"fmt"
	"io"
	"os"
	"time"
)

type FileOptions struct {
	Filename string `env:"FILENAME" flag:"filename" desc:"filename to read from"`

	Repeat      bool          `env:"FILE_REPEAT" flag:"repeat" desc:"if set, the file will be offered repeatedly"`
	RepeatDelay time.Duration `env:"FILE_REPEAT_DELAY" flag:"repeat-delay" desc:"delay between repetitions"`
}

func OpenFile(opts *FileOptions) (io.ReadCloser, error) {
	if opts.Repeat {
		data, err := os.ReadFile(opts.Filename)
		if err != nil {
			return nil, err
		}

		return &dataRepeater{
			data:   data,
			ticker: time.NewTicker(opts.RepeatDelay),
		}, nil
	}

	file, err := os.Open(opts.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", opts.Filename, err)
	}

	return file, nil
}

type dataRepeater struct {
	data []byte

	ticker *time.Ticker
}

func (r *dataRepeater) Read(p []byte) (int, error) {
	if r.ticker == nil {
		return 0, io.EOF
	}

	<-r.ticker.C

	n := copy(p, r.data)

	return n, nil
}

func (r *dataRepeater) Close() error {
	r.ticker.Stop()
	r.ticker = nil

	return nil
}
