// Package vault implements a koanf.Provider for Hashicorp Vault KV secrets engine
// and provides it to koanf to be parsed by a koanf.Parser.
package vault

import (
	"errors"
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"
)

type Config struct {
	// Vault server address
	Address string

	// Vault static token
	Token string

	// Secret data path
	Path string

	// Internal HTTP client timeout
	Timeout time.Duration
}

type Vault struct {
	client *api.Client
	cfg    Config
}

func Provider(cfg Config) *Vault {
	httpClient := &http.Client{Timeout: cfg.Timeout}
	client, err := api.NewClient(&api.Config{Address: cfg.Address, HttpClient: httpClient})
	if err != nil {
		return nil
	}
	client.SetToken(cfg.Token)

	return &Vault{client: client, cfg: cfg}
}

func (r *Vault) ReadBytes() ([]byte, error) {
	return nil, errors.New("vault provider does not support this method")
}

func (r *Vault) Read() (map[string]interface{}, error) {
	secret, err := r.client.Logical().Read(r.cfg.Path)
	if err != nil {
		return nil, err
	}

	return secret.Data, nil
}
