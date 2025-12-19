// Package kiln implements a koanf.Provider that reads encrypted environment
// variables from kiln-encrypted files as conf maps.
package kiln

import (
	"errors"
	"strings"

	"github.com/thunderbottom/kiln/pkg/kiln"
)

// Kiln implements a kiln encrypted environment variables provider.
type Kiln struct {
	configPath string
	keyPath    string
	file       string
	prefix     string
	transform  func(key, value string) (string, any)
}

// Opt holds optional configuration for the kiln provider.
type Opt struct {
	// If specified (case-sensitive), only env vars beginning with
	// the prefix are captured. eg: "APP_"
	Prefix string

	// TransformFunc is an optional callback that takes an environment
	// variable's string name and value, runs arbitrary transformations
	// on them and returns a transformed string key and value of any type.
	// Common usecase are stripping prefixes from keys, lowercasing variable names,
	// replacing _ with . etc. Eg: APP_DB_HOST -> db.host
	// If the returned variable name is an empty string (""), it is ignored altogether.
	TransformFunc func(k, v string) (string, any)
}

// Provider returns a kiln provider that returns a nested map[string]any
// of decrypted environment variables from the specified kiln environment file.
//
// configPath is the path to the kiln.toml configuration file.
// keyPath is the path to the private key file (age or SSH format).
// If keyPath is empty, the provider will attempt to auto-discover a compatible
// private key from standard locations (~/.kiln/kiln.key, ~/.ssh/id_ed25519, etc.).
// file is the environment file name as defined in kiln.toml (e.g., "production", "staging").
// opt provides optional configuration for key filtering and transformation.
//
// The provider handles decryption, access control validation, and secure memory
// cleanup automatically. All operations respect the role-based access controls
// defined in the kiln.toml configuration file.
func Provider(configPath, keyPath, file string, opt Opt) *Kiln {
	return &Kiln{
		configPath: configPath,
		keyPath:    keyPath,
		file:       file,
		prefix:     opt.Prefix,
		transform:  opt.TransformFunc,
	}
}

// ReadBytes is not supported by the kiln provider.
func (k *Kiln) ReadBytes() ([]byte, error) {
	return nil, errors.New("kiln provider does not support this method")
}

// Read reads all available encrypted environment variables from the specified
// kiln environment file into a key:value map and returns it.
func (k *Kiln) Read() (map[string]any, error) {
	cfg, err := kiln.LoadConfig(k.configPath)
	if err != nil {
		return nil, err
	}

	keyPath := k.keyPath
	if keyPath == "" {
		if discovered, discErr := kiln.DiscoverPrivateKey(); discErr == nil {
			keyPath = discovered
		} else {
			return nil, discErr
		}
	}

	identity, err := kiln.NewIdentityFromKey(keyPath)
	if err != nil {
		return nil, err
	}
	defer identity.Cleanup()

	vars, cleanup, err := kiln.GetAllEnvironmentVars(identity, cfg, k.file)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	out := make(map[string]any)
	for key, value := range vars {
		if k.prefix != "" {
			if !strings.HasPrefix(key, k.prefix) {
				continue
			}
		}

		// value is []bytes, so we need to convert it to
		// string and accept any for the transformation
		val := any(string(value))

		// Apply key transformation, if there's a callback.
		if k.transform != nil {
			key, transformedValue := k.transform(key, string(value))

			// If the callback blanked the key, omit it.
			if key == "" {
				continue
			}

			out[key] = transformedValue
		}

		out[key] = val
	}

	return out, nil
}
