// Package fs implements a koanf.Provider that reads raw bytes
// from given fs.FS to be used with a koanf.Parser to parse
// into conf maps.

//go:build go1.16
// +build go1.16

package fs

import (
	"errors"
	"io"
	"io/fs"
)

// FS implements an fs.FS provider.
type FS struct {
	fs   fs.FS
	path string
}

// Provider returns an fs.FS provider.
func Provider(fs fs.FS, filepath string) *FS {
	return &FS{fs: fs, path: filepath}
}

// ReadBytes reads the contents of given filepath from fs.FS and returns the bytes.
func (f *FS) ReadBytes() ([]byte, error) {
	fd, err := f.fs.Open(f.path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	return io.ReadAll(fd)
}

// Read is not supported by the fs.FS provider.
func (f *FS) Read() (map[string]any, error) {
	return nil, errors.New("fs.FS provider does not support this method")
}
