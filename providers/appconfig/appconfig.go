// Package appconfig implements a koanf.Provider for AWS AppConfig
// and provides it to koanf to be parsed by a koanf.Parser.
package appconfig

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Config holds the AWS AppConfig Configuration.
type Config struct {
	// The AWS AppConfig Application to get. Specify either the application
	// name or the application ID.
	Application string

	// The Client ID for the AppConfig. Enables AppConfig to deploy the
	// configuration in intervals, as defined in the deployment strategy.
	ClientID string

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
}

// AppConfig implements an AWS AppConfig provider.
type AppConfig struct {
	client *appconfig.Client
	config Config
	input  appconfig.GetConfigurationInput
}

// Provider returns an AWS AppConfig provider.
func Provider(cfg Config) *AppConfig {
	c, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil
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
	client := appconfig.NewFromConfig(c)

	return &AppConfig{client: client, config: cfg}
}

// ReadBytes returns the raw bytes for parsing.
func (ac *AppConfig) ReadBytes() ([]byte, error) {
	ac.input = appconfig.GetConfigurationInput{
		Application:   &ac.config.Application,
		ClientId:      &ac.config.ClientID,
		Configuration: &ac.config.Configuration,
		Environment:   &ac.config.Environment,
	}
	if ac.config.ClientConfigurationVersion != "" {
		ac.input.ClientConfigurationVersion = &ac.config.ClientConfigurationVersion
	}

	conf, err := ac.client.GetConfiguration(context.TODO(), &ac.input)
	if err != nil {
		return nil, err
	}

	// Set the response configuration version as the current configuration version.
	// Useful for Watch().
	ac.input.ClientConfigurationVersion = conf.ConfigurationVersion

	return conf.Content, nil
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
			conf, err := ac.client.GetConfiguration(context.TODO(), &ac.input)
			if err != nil {
				cb(nil, err)
				break loop
			}

			// Check if the the configuration has been updated.
			if len(conf.Content) == 0 {
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
