// Package vault implements a koanf.Provider for Hashicorp Vault KV secrets engine
// and provides it to koanf to be parsed by a koanf.Parser.
package vault

import (
	"errors"
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/knadh/koanf/maps"
)

type Config struct {
	// Vault server address
	Address string

	// Vault static token
	Token string

	// Secret data path
	Path string

	// If FlatPaths is true, then the loaded configuration is not split into
	// hierarchical maps based on the delimiter. The keys including the delimiter,
	// eg: app.db.name stays as-is in the confmap.
	FlatPaths bool

	// Delim is the delimiter to use
	// when specifying config key paths, for instance a . for `parent.child.key`
	// or a / for `parent/child/key`.
	Delim string

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

	if !r.cfg.FlatPaths {
		data := maps.Unflatten(secret.Data, r.cfg.Delim)

		return data, nil
	}

	return secret.Data, nil
}
