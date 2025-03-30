package main

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/knadh/koanf/providers/azkeyvault"
	"github.com/knadh/koanf/v2"
	"log"
	"os"
)

// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
var k = koanf.New(".")

func main() {
	clientId := os.Getenv("ARM_CLIENT_ID")
	clientSecret := os.Getenv("ARM_CLIENT_SECRET")
	tenantId := os.Getenv("ARM_TENANT_ID")

	tokenCred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, nil)
	if err != nil {
		log.Fatalf("error creating token credential: %v", err)
	}

	config := azkeyvault.Config{
		KeyVaultUrl:     "https://mykeyvault.vault.azure.net",
		TokenCredential: tokenCred,
	}

	provider, err := azkeyvault.Provider(config)
	if err != nil {
		log.Fatalf("Failed to instantiate azure key vault provider: %v", err)
	}

	if err := k.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	fmt.Println("database's host is = ", k.String("database.host"))
	fmt.Println("database's port is = ", k.Int("database.port"))
}
