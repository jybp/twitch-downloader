package twitchdl

import (
	"io"

	"github.com/pkg/errors"
)

// merger merges the "downloads" into a single io.Reader.
type merger struct {
	downloads []downloadFunc

	index   int
	current io.ReadCloser
	err     error
}

func (r *merger) next() error {
	if r.index >= len(r.downloads) {
		r.current = nil
		return nil
	}
	var err error
	r.current, err = r.downloads[r.index]()
	r.index++
	return err
}

func (r *merger) Read(p []byte) (int, error) {
	for {
		if r.err != nil {
			return 0, r.err
		}
		if r.current != nil {
			n, err := r.current.Read(p)
			if err == io.EOF {
				err = r.current.Close()
				r.current = nil
			}
			return n, errors.WithStack(err)
		}
		if err := r.next(); err != nil {
			return 0, err
		}
		if r.current == nil {
			return 0, io.EOF
		}
	}
}
