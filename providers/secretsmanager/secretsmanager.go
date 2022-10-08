// Package secretsmanager implements a koanf.Provider for AWS SecretsManager
// and provides it to koanf to be parsed by a koanf.Parser.
package secretsmanager

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/knadh/koanf/maps"
	"github.com/knadh/koanf/parsers/json"
)

// Config holds the AWS SecretsManager Configuration.
type Config struct {
	// The AWS SecretsManager Delim that might be used 
	// delim string
	Delim string

	// The SecretsManager name or arn to fetch
	// name or the secret ID.
	SecretId string

	// The type of values secre value set, it can only be string or map. 
	// if the value is type of app, each key is unfallten to create new 
	// single var like: parent: {"child": "value"} -> parent.child = value
	Type string

	// The SecretsManager Configuration Version to fetch. Specifying a VersionId
	// ensures that the configuration is only fetched if it is updated. If not specified,
	// the latest available configuration is fetched always.
	// Setting this to the latest configuration version will return an empty slice of bytes.
	VersionId *string

	// The AWS Access Key ID to use. This value is fetched from the environment
	// if not specified.
	AWSAccessKeyID string

	// The AWS Secret Access Key to use. This value is fetched from the environment
	// if not specified.
	AWSSecretAccessKey string

	// The AWS IAM Role ARN to use. Useful for access requiring IAM AssumeRole.
	AWSRoleARN string

	// The AWS Region to use. This value is fetched from teh environment if not specified.
	AWSRegion string

	// Time interval at which the watcher will refresh the configuration.
	// Defaults to 3600 seconds.
	WatchInterval time.Duration
}

// SMConfig implements an AWS SecretsManager provider.
type SMConfig struct {
	client  *secretsmanager.Client
	config  Config
	input   secretsmanager.GetSecretValueInput
	cb 		func(s string) string
}

// Provider returns an AWS SecretsManager provider.
func Provider(cfg Config, cb func(s string) string) *SMConfig {
	// load the default config
	c, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil
	}

	// check inputs and set
	if cfg.Delim == "" {
		cfg.Delim = "_"
	}
	if cfg.AWSRegion != "" {
		c.Region = cfg.AWSRegion
	}

	// Check if AWS Access Key ID and Secret Key are specified.
	if cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" {
		c.Credentials = credentials.NewStaticCredentialsProvider(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, "")
	}

	// Check if AWS Role ARN is present.
	if cfg.AWSRoleARN != "" {
		stsSvc := sts.NewFromConfig(c)
		credentials := stscreds.NewAssumeRoleProvider(stsSvc, cfg.AWSRoleARN)
		c.Credentials = aws.NewCredentialsCache(credentials)
	}
	client := secretsmanager.NewFromConfig(c)

	return &SMConfig{client: client, config: cfg, cb: cb}
}

// ProviderWithClient returns an AWS SecretsManager provider
// using an existing AWS SecretsManager client.
func ProviderWithClient(cfg Config, cb func(s string) string, client *secretsmanager.Client) *SMConfig {
	return &SMConfig{client: client, config: cfg, cb: cb}
}

// Read is not supported by the SecretsManager provider.
func (sm *SMConfig) Read() (map[string]interface{}, error)  {

	// check if secretId is provided
	if sm.config.SecretId == "" {
		return nil, errors.New("no secret id  provided")
	}

	// set secretsmanger input
	sm.input = secretsmanager.GetSecretValueInput{
		SecretId: aws.String(sm.config.SecretId),
	}

	// check if latest version exist
	if sm.config.VersionId != nil {
		sm.input.VersionId = sm.config.VersionId
	}

	// get secret value
	conf, err := sm.client.GetSecretValue(context.TODO(), &sm.input)
	if err != nil {
		return nil, err
	}

	mp := make(map[string]interface{})

	// check if secret is set as string
	if conf.SecretString != nil {
		key := *conf.Name
		// transform key id transformer provided
		if sm.cb != nil {
			key = sm.cb(key)
		}
		if key == "" {
			return nil, errors.New("transformed key has become null")
		}
		// set key value
		mp[key] = *conf.SecretString
	}
	
	// if value is set as map it will unfaltten 
	if sm.config.Type == "map" {
		// parse secret value as map if type is set as map
		valueMap, err := json.Parser().Unmarshal([]byte(*conf.SecretString))
		if err != nil {
			return nil, errors.New("unable to unmarshal value as obj")
		}
		// modify each value
		for k, v := range valueMap {
			updated_key := *conf.Name + sm.config.Delim + k
			// transform key id transformer provided
			if sm.cb != nil {
				updated_key = sm.cb(updated_key)
			}
			// If the callback blanked the key, it should be omitted
			if updated_key == "" {
				continue
			}
			mp[updated_key] = v
		}
	}


	// Set the response configuration version as the current configuration version.
	// Useful for Watch().
	sm.config.VersionId = conf.VersionId

	return maps.Unflatten(mp, sm.config.Delim), nil
}


// ReadBytes returns the raw bytes for parsing.
func (sm *SMConfig) ReadBytes() ([]byte, error) {
	// shoud implement for SecretBinary. maybe in future 
	return nil, errors.New("secretsmanager provider does not support this method")
}

// Watch polls AWS AppConfig for configuration updates.
func (sm *SMConfig) Watch(cb func(event interface{}, err error)) error {
	if sm.config.WatchInterval == 0 {
		// Set default watch interval to 3600 seconds. to reduce cost
		sm.config.WatchInterval = 3600 * time.Second
	}

	go func() {
	loop:
		for {
			conf, err := sm.client.GetSecretValue(context.TODO(), &sm.input)
			if err != nil {
				cb(nil, err)
				break loop
			}

			// Check if the the configuration has been updated.
			if *conf.VersionId == *sm.config.VersionId {
				// Configuration is not updated and we have the latest version.
				// Sleep for WatchInterval and retry watcher.
				time.Sleep(sm.config.WatchInterval)
				continue
			}

			// Trigger event.
			cb(nil, nil)
		}
	}()

	return nil
}
