// Package appconfig implements a koanf.Provider for AWS AppConfig
// and provides it to koanf to be parsed by a koanf.Parser.
package appconfig

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/appconfigdata"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Config holds the AWS AppConfig Configuration.
type Config struct {
	// The AWS AppConfig Application to get. Specify either the application
	// name or the application ID.
	Application string

	// The AppConfig configuration to fetch. Specify either the configuration
	// name or the configuration ID.
	Configuration string

	// The AppConfig environment to get. Specify either the environment
	// name or the environment ID.
	Environment string

	// The AppConfig Configuration Version to fetch. Specifying a ClientConfigurationVersion
	// ensures that the configuration is only fetched if it is updated. If not specified,
	// the latest available configuration is fetched always.
	// Setting this to the latest configuration version will return an empty slice of bytes.
	ClientConfigurationVersion string

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
	// Defaults to 60 seconds.
	WatchInterval time.Duration

	// (Optional) Sets a constraint on a session. If you specify a value of, for example, 60 seconds, then the client
	// that established the session can't call GetLatestConfiguration more frequently than every 60 seconds.
	// Valid Range: Minimum value of 15. Maximum value of 86400.
	RequiredMinimumPollIntervalInSeconds *int32
}

// AppConfig implements an AWS AppConfig provider.
type AppConfig struct {
	client *appconfigdata.Client
	config Config
	token  *string
}

// Provider returns an AWS AppConfig provider.
func Provider(cfg Config) (*AppConfig, error) {
	c, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
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
		roleCredentials := stscreds.NewAssumeRoleProvider(stsSvc, cfg.AWSRoleARN)
		c.Credentials = aws.NewCredentialsCache(roleCredentials)
	}

	client := appconfigdata.NewFromConfig(c)

	return &AppConfig{client: client, config: cfg}, nil
}

// ProviderWithClient returns an AWS AppConfig provider
// using an existing AWS appconfig client.
func ProviderWithClient(cfg Config, client *appconfigdata.Client) *AppConfig {
	return &AppConfig{client: client, config: cfg}
}

func (ac *AppConfig) getLatestConfiguration() ([]byte, error) {
	ctx := context.TODO()

	if ac.token == nil {
		// We don't need to save off this input anymore
		input := &appconfigdata.StartConfigurationSessionInput{
			ApplicationIdentifier:                &ac.config.Application,
			ConfigurationProfileIdentifier:       &ac.config.Configuration,
			EnvironmentIdentifier:                &ac.config.Environment,
			RequiredMinimumPollIntervalInSeconds: ac.config.RequiredMinimumPollIntervalInSeconds,
		}

		startOutput, err := ac.client.StartConfigurationSession(ctx, input)

		if err != nil {
			return nil, err
		}

		ac.token = startOutput.InitialConfigurationToken
	}

	configInput := &appconfigdata.GetLatestConfigurationInput{
		ConfigurationToken: ac.token,
	}

	configOutput, err := ac.client.GetLatestConfiguration(ctx, configInput)

	if err != nil {
		return nil, err
	}

	ac.token = configOutput.NextPollConfigurationToken

	return configOutput.Configuration, nil
}

// ReadBytes returns the raw bytes for parsing.
func (ac *AppConfig) ReadBytes() ([]byte, error) {
	return ac.getLatestConfiguration()
}

// Read is not supported by the appconfig provider.
func (ac *AppConfig) Read() (map[string]interface{}, error) {
	return nil, errors.New("appconfig provider does not support this method")
}

// Watch polls AWS AppConfig for configuration updates.
func (ac *AppConfig) Watch(cb func(event interface{}, err error)) error {
	if ac.config.WatchInterval == 0 {
		// Set default watch interval to 60 seconds.
		ac.config.WatchInterval = 60 * time.Second
	}

	go func() {
	loop:
		for {
			conf, err := ac.getLatestConfiguration()
			if err != nil {
				cb(nil, err)
				break loop
			}

			// Check if the configuration has been updated.
			if len(conf) == 0 {
				// Configuration is not updated and we have the latest version.
				// Sleep for WatchInterval and retry watcher.
				time.Sleep(ac.config.WatchInterval)
				continue
			}

			// Trigger event.
			cb(nil, nil)
		}
	}()

	return nil
}
