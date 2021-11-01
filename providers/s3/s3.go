// Package s3 implements a koanf.Provider that takes a []byte slice
// and provides it to koanf to be parsed by a koanf.Parser.
package s3

import (
	"errors"
	"io/ioutil"

	"github.com/rhnvrm/simples3"
)

// Config for the provider.
type Config struct {
	// AWS Access Key
	AccessKey string

	// AWS Secret Key
	SecretKey string

	// AWS region
	Region string

	// Bucket Name
	Bucket string

	// Object Key
	ObjectKey string

	// Optional: Custom S3 compatible endpoint
	Endpoint string
}

// S3 implements a s3 provider.
type S3 struct {
	s3  *simples3.S3
	cfg Config
}

// Provider returns a provider that takes a simples3 config.
func Provider(cfg Config) *S3 {
	s3 := simples3.New(cfg.Region, cfg.AccessKey, cfg.SecretKey)
	s3.SetEndpoint(cfg.Endpoint)

	return &S3{s3: s3, cfg: cfg}
}

// ReadBytes reads the contents of a file on s3 and returns the bytes.
func (r *S3) ReadBytes() ([]byte, error) {
	resp, err := r.s3.FileDownload(simples3.DownloadInput{
		Bucket:    r.cfg.Bucket,
		ObjectKey: r.cfg.ObjectKey,
	})
	if err != nil {
		return nil, err
	}

	defer resp.Close()

	data, err := ioutil.ReadAll(resp)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Read returns the raw bytes for parsing.
func (r *S3) Read() (map[string]interface{}, error) {
	return nil, errors.New("s3 provider does not support this method")
}

// Watch is not supported.
func (r *S3) Watch(cb func(event interface{}, err error)) error {
	return errors.New("s3 provider does not support this method")
}
