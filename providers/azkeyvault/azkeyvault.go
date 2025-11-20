package azkeyvault

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/knadh/koanf/maps"
)

type Config struct {
	// KeyVaultUrl key vault url
	KeyVaultUrl string

	// Delim is the delimiter to use
	// when specifying config key paths, for instance a . for `parent.child.key`
	// or a / for `parent/child/key`.
	Delim string

	// TokenCredential Token credential to connect to the Azure Keyvault.
	// It can be created using client Id with client secret or client certificate.
	TokenCredential azcore.TokenCredential

	// If FlatPaths is true, then the loaded configuration is not split into
	// hierarchical maps based on the delimiter. The keys including the delimiter,
	// eg: app.db.name stays as-is in the confmap.
	FlatPaths bool
}

type AzureKeyVault struct {
	kvclient *azsecrets.Client
	config   Config
}

func Provider(config Config) (*AzureKeyVault, error) {
	kvClient, err := azsecrets.NewClient(config.KeyVaultUrl, config.TokenCredential, nil)
	if err != nil {
		return nil, err
	}

	return &AzureKeyVault{kvclient: kvClient, config: config}, nil
}

func (kv *AzureKeyVault) ReadBytes() ([]byte, error) {
	return nil, errors.New("azure key vault provider does not support this method")
}

func (kv *AzureKeyVault) Read() (map[string]any, error) {
	ctx := context.Background()
	secrets := make(map[string]any)

	// Get all the secrets from the KV
	pager := kv.kvclient.NewListSecretPropertiesPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, secret := range page.Value {
			sName := secret.ID.Name()
			SValue, err := kv.kvclient.GetSecret(ctx, sName, "", nil)

			if err != nil {
				return nil, fmt.Errorf("failed to get the secret value for %s, Error: %w", sName, err)
			}

			secrets[sName] = *SValue.Value
		}
	}

	if kv.config.Delim != "" && !kv.config.FlatPaths {
		data := maps.Unflatten(secrets, kv.config.Delim)
		return data, nil
	}

	return secrets, nil
}
