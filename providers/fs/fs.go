// Package fs implements a koanf.Provider that reads raw bytes
// from given fs.FS to be used with a koanf.Parser to parse
// into conf maps.

// +build go1.16

package fs

import (
	"errors"
	"io/fs"
	"io/ioutil"
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

	return ioutil.ReadAll(fd)
}

// Read is not supported by the fs.FS provider.
func (f *FS) Read() (map[string]interface{}, error) {
	return nil, errors.New("fs.FS provider does not support this method")
}

// Watch is not supported by the fs.FS provider.
func (f *FS) Watch(cb func(event interface{}, err error)) error {
	return errors.New("fs.FS provider does not support this method")
}
