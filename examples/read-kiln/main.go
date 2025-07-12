package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/kiln"
	"github.com/knadh/koanf/v2"
)

var k = koanf.New(".")

// Run this example:
//
//	kiln init key
//	kiln init config --path="./examples/read-kiln/kiln.toml" --recipients $(whoami)=$(cat ~/.kiln/kiln.key.pub)
//	export KILN_CONFIG_FILE=./examples/read-kiln/kiln.toml
//	kiln set API_KEY super-secret-dev-key
//	kiln set API_HOST example.com
//	kiln set DB_HOST postgres://user@localhost:5432
//	go run ./examples/read-kiln-environment
//
// This example demonstrates loading configuration from multiple sources:
// 1. A JSON config file for defaults
// 2. Encrypted kiln environment variables that override defaults
//
// The kiln provider decrypts environment variables from encrypted files
// and applies transformations similar to the env provider.
//
// Example encrypted environment file (.kiln.env):
//
//	API_KEY=super-secret-dev-key
//	API_HOST=example.com
func main() {
	// Load JSON config for default values.
	if err := k.Load(file.Provider("mock/mock.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	// Load encrypted environment variables from kiln and merge into config.
	// The kiln provider decrypts variables from the specified environment file
	// and applies optional transformations for seamless integration.
	//
	// In this example:
	// - Only variables with "API_" prefix are loaded
	// - Keys are transformed to lowercase with dots for nesting
	// - The prefix "api." is stripped
	// - API_KEY becomes "key", API_HOST becomes "host"
	//
	// The provider automatically handles:
	// - Key discovery from standard locations (~/.kiln/kiln.key, ~/.ssh/id_*)
	// - Decryption using age or SSH keys
	// - Access control validation per kiln.toml
	// - Secure memory cleanup after operations
	if err := k.Load(kiln.Provider("kiln.toml", "", "default", kiln.Opt{
		Prefix: "API_",
		TransformFunc: func(key, value string) (string, any) {
			// Strip common prefixes and normalize to lowercase with dots
			key = strings.ToLower(key)
			key = strings.ReplaceAll(key, "_", ".")

			// Remove prefixes for cleaner config keys
			if strings.HasPrefix(key, "api.") {
				key = strings.TrimPrefix(key, "api.")
			}

			return key, value
		},
	}), nil); err != nil {
		log.Fatalf("error loading kiln config: %v", err)
	}

	// Display the merged configuration
	fmt.Println("api key is =", k.String("key"))
	fmt.Println("api host is =", k.String("host"))
	fmt.Println("Parent name from JSON =", k.String("parent1.name"))
}
