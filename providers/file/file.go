// Package file implements a koanf.Provider that reads raw bytes
// from files on disk to be used with a koanf.Parser to parse
// into conf maps.
package file

import (
	"errors"
	"io/ioutil"
)

// File implements a File provider.
type File struct {
	path string
}

// Provider returns a file provider.
func Provider(path string) *File {
	return &File{path: path}
}

// ReadBytes reads the contents of a file on disk and returns the bytes.
func (f *File) ReadBytes() ([]byte, error) {
	return ioutil.ReadFile(f.path)
}

// Read is not supported by the file provider.
func (f *File) Read() (map[string]interface{}, error) {
	return nil, errors.New("file provider does not support this method")
}
