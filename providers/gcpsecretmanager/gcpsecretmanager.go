// Package gcpsecretmanager implements a koanf.Provider for Google Cloud Secret Manager.
package gcpsecretmanager

import (
	"context"
	"errors"
	"fmt"
	"path"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/knadh/koanf/maps"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config represents a GCP Secret Manager provider configuration.
type Config struct {
	// Project is the ID or number of the GCP project to read secrets from.
	Project string

	// Filter is an optional Secret Manager list filter, for instance
	// `labels.app=myapp` or `name:myapp_`. See
	// https://cloud.google.com/secret-manager/docs/filtering
	Filter string

	// Delim is the delimiter to use when specifying config key paths.
	// Secret IDs may only contain letters, digits, `_` and `-`, so use one of
	// those (or a Callback) to get nested keys, eg: `db_password` with `_`
	// becomes `db.password`.
	Delim string

	// If FlatPaths is true, then the loaded configuration is not split into
	// hierarchical maps based on the delimiter.
	FlatPaths bool

	// Callback is an optional callback that takes a (key, value) with the
	// secret ID and value and allows you to modify both. If the callback
	// returns an empty key, the secret is ignored.
	Callback func(key, value string) (string, any)
}

// SecretManager implements a GCP Secret Manager provider.
type SecretManager struct {
	client *secretmanager.Client
	opts   []option.ClientOption
	config Config
}

// Provider returns a SecretManager provider. The client is created on each
// Read with Application Default Credentials and the given options.
func Provider(config Config, opts ...option.ClientOption) *SecretManager {
	return &SecretManager{config: config, opts: opts}
}

// ProviderWithClient returns a SecretManager provider that uses an existing
// client. The caller is responsible for closing it.
func ProviderWithClient(config Config, client *secretmanager.Client) *SecretManager {
	return &SecretManager{config: config, client: client}
}

// ReadBytes is not supported by the SecretManager provider.
func (s *SecretManager) ReadBytes() ([]byte, error) {
	return nil, errors.New("gcpsecretmanager provider does not support this method")
}

// Read returns the latest enabled version of every secret in the project
// (that matches the filter) as a config map keyed by secret ID.
func (s *SecretManager) Read() (map[string]any, error) {
	if s.config.Project == "" {
		return nil, errors.New("gcpsecretmanager: project is required")
	}

	ctx := context.Background()
	client := s.client
	if client == nil {
		c, err := secretmanager.NewClient(ctx, s.opts...)
		if err != nil {
			return nil, err
		}
		defer c.Close()
		client = c
	}

	mp := make(map[string]any)
	it := client.ListSecrets(ctx, &secretmanagerpb.ListSecretsRequest{
		Parent: "projects/" + s.config.Project,
		Filter: s.config.Filter,
	})
	for {
		secret, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}

		resp, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
			Name: secret.GetName() + "/versions/latest",
		})
		if err != nil {
			// The secret has no versions, or its latest version is disabled or destroyed.
			if c := status.Code(err); c == codes.NotFound || c == codes.FailedPrecondition {
				continue
			}
			return nil, fmt.Errorf("gcpsecretmanager: accessing %s: %w", secret.GetName(), err)
		}

		key := path.Base(secret.GetName())
		value := string(resp.GetPayload().GetData())
		if s.config.Callback == nil {
			mp[key] = value
			continue
		}
		k, v := s.config.Callback(key, value)
		if k == "" {
			continue
		}
		mp[k] = v
	}

	if s.config.Delim != "" && !s.config.FlatPaths {
		return maps.Unflatten(mp, s.config.Delim), nil
	}
	return mp, nil
}
