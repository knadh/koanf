// Package consul implements a koanf.Provider that takes a []byte slice
// from Consul KV storage
// and provides it to koanf to be parsed by a koanf.Parser.
package consul

import (
	"errors"

	"github.com/hashicorp/consul/api"
)

// Config for the provider.
type Config struct {
	// Consul Access Key
	Token string

	// Consul endpoint
	Address string

	//Key to get from KV storage
	Key string
}

// Consul implements a consul provider.
type Consul struct {
	consul *api.Client
	cfg    Config
}

// Provider returns a provider that takes a Consul config.
func Provider(cfg Config) *Consul {
	consulClient, err := api.NewClient(&api.Config{Address: cfg.Address, Token: cfg.Token})
	if err != nil {
		return nil
	}
	return &Consul{consul: consulClient, cfg: cfg}
}

// ReadBytes reads the contents of a KV from Consul and returns the bytes.
func (r *Consul) ReadBytes() ([]byte, error) {
	kv := r.consul.KV()
	pair, _, err := kv.Get(r.cfg.Key, nil)

	if err != nil {
		return nil, err
	}

	return pair.Value, nil
}

// Read returns the raw bytes for parsing.
func (r *Consul) Read() (map[string]interface{}, error) {
	return nil, errors.New("Consul provider does not support this method")
}
