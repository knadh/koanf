package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/knadh/koanf/providers/vault/v2"
	"github.com/knadh/koanf/v2"
)

// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
var k = koanf.New(".")

func main() {
	provider, err := vault.Provider(vault.Config{
		Address: os.Getenv("VAULT_ADDRESS"),
		Token:   os.Getenv("VAULT_TOKEN"),
		Path:    "secret/data/my-app",
		Timeout: 10 * time.Second,

		// If this is set to false, then `data` and `metadata` keys
		// from Vault are fetched. All config is then accessed as
		// k.String("data.YOUR_KEY") etc. instead of k.String("YOUR_KEY").
		ExcludeMeta: true,
	})
	if err != nil {
		log.Fatalf("Failed to instantiate vault provider: %v", err)
	}
	// Load mapped config from Vault storage.
	if err := k.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	fmt.Println("database's host is = ", k.String("database.host"))
	fmt.Println("database's port is = ", k.Int("database.port"))
}
