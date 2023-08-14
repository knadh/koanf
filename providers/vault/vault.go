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

	// ExcludeMeta states whether the secret should be returned with its metadata.
	// If ExcludeMeta is true, no metadata will be returned, and the data can be
	// accessed as `k.String("key")`. If set to false, the value for data `key`
	// and the metadata `version` can be accessed as `k.String("data.key")` and
	// `k.Int("metadata.version")`.
	ExcludeMeta bool
}

type Vault struct {
	client *api.Client
	cfg    Config
}

// Provider returns a provider that takes a Vault config.
func Provider(cfg Config) *Vault {
	httpClient := &http.Client{Timeout: cfg.Timeout}
	client, err := api.NewClient(&api.Config{Address: cfg.Address, HttpClient: httpClient})
	if err != nil {
		return nil
	}
	client.SetToken(cfg.Token)

	return &Vault{client: client, cfg: cfg}
}

// ReadBytes is not supported by the vault provider.
func (r *Vault) ReadBytes() ([]byte, error) {
	return nil, errors.New("vault provider does not support this method")
}

// Read fetches the configuration from the source and returns a nested config map.
func (r *Vault) Read() (map[string]interface{}, error) {
	secret, err := r.client.Logical().Read(r.cfg.Path)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		return nil, errors.New("vault provider fetched no data")
	}

	s := secret.Data
	if r.cfg.ExcludeMeta {
		s = secret.Data["data"].(map[string]interface{})
	}

	// Unflatten only when a delimiter is specified
	if !r.cfg.FlatPaths && r.cfg.Delim != "" {
		data := maps.Unflatten(s, r.cfg.Delim)

		return data, nil
	}

	return s, nil
}
