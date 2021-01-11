package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/vault"
)

// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
var k = koanf.New(".")

func main() {
	provider := vault.Provider(vault.Config{
		Address: os.Getenv("VAULT_ADDRESS"),
		Token:   os.Getenv("VAULT_TOKEN"),
		Path:    "secret/data/my-app",
		Timeout: 10 * time.Second,
	})
	// Load mapped config from Vault storage.
	if err := k.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	fmt.Println("database's host is = ", k.String("database.host"))
	fmt.Println("database's port is = ", k.Int("database.port"))
}
