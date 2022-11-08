// Package parameterstore implements a koanf.Provider for AWS parameterstore
// and provides it to koanf to be parsed by a koanf.Parser.
package parameterstore

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/knadh/koanf/maps"
	"github.com/knadh/koanf/parsers/json"
)

// Config holds the AWS parameterstore Configuration.
type Config struct {
	// The AWS parameterstore Delim that might be used
	// delim string
	Delim string

	// The parameterstore name  to fetch
	// name of the parameter.
	Name string

	// The type of values secre value set, it can only be string or map.
	// if the value is type of app, each key is unfallten to create new
	// single var like: parent: {"child": "value"} -> parent.child = value
	Type string

	// The ParameterStore Configuration Version to fetch. Specifying a VersionId
	// ensures that the configuration is only fetched if it is updated. If not specified,
	// the latest available configuration is fetched always.
	// Setting this to the latest configuration version will return an empty slice of bytes.
	VersionId string

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
	// Defaults to 600 seconds.
	WatchInterval time.Duration
}

// PSConfig implements an AWS ParameterStore provider.
type ParameterStore struct {
	client *ssm.Client
	config Config
	input  ssm.GetParameterInput
	cb     func(s string) string
}

// Provider returns an AWS ParameterStore provider.
func Provider(cfg Config, cb func(s string) string) *ParameterStore {
	c, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil
	}

	// Set defaults.
	if cfg.Delim == "" {
		cfg.Delim = "_"
	}
	if cfg.AWSRegion != "" {
		c.Region = cfg.AWSRegion
	}
	if cfg.AWSAccessKeyID != "" || cfg.AWSSecretAccessKey != "" {
		c.Credentials = credentials.NewStaticCredentialsProvider(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, "")
	}

	// Check if AWS Role ARN is present.
	if cfg.AWSRoleARN != "" {
		var (
			stsSvc      = sts.NewFromConfig(c)
			credentials = stscreds.NewAssumeRoleProvider(stsSvc, cfg.AWSRoleARN)
		)
		c.Credentials = aws.NewCredentialsCache(credentials)
	}
	client := ssm.NewFromConfig(c)

	return &ParameterStore{client: client, config: cfg, cb: cb}
}

// ProviderWithClient returns an AWS ParameterStore provider
// using an existing AWS parameterstore client.
func ProviderWithClient(cfg Config, cb func(s string) string, client *ssm.Client) *ParameterStore {
	return &ParameterStore{client: client, config: cfg, cb: cb}
}

// Read is not supported by the ParameterStore provider.
func (ps *ParameterStore) Read() (map[string]interface{}, error) {
	if ps.config.Name == "" {
		return nil, errors.New("no parameter name provided")
	}

	ps.input = ssm.GetParameterInput{
		Name: aws.String(ps.config.Name),
	}

	// If the latest version exists, then update name is "Name": "name:version".
	if ps.config.VersionId != "" {
		ps.input.Name = aws.String(*ps.input.Name + ":" + ps.config.VersionId)
	}

	// Fetch the params.
	conf, err := ps.client.GetParameter(context.TODO(), &ps.input)
	if err != nil {
		return nil, err
	}

	mp := make(map[string]interface{})

	if (conf.Parameter.Type == types.ParameterTypeString) || (conf.Parameter.Type == types.ParameterTypeStringList) {
		key := *conf.Parameter.Name

		// Optionally transform the key.
		if ps.cb != nil {
			key = ps.cb(key)
		}
		if key == "" {
			return nil, errors.New("transformed key is empty")
		}

		mp[key] = *conf.Parameter.Value
	}

	// Unflatten map.
	if ps.config.Type == "map" {
		mp = make(map[string]interface{})

		// Parse secret as map.
		valueMap, err := json.Parser().Unmarshal([]byte(*conf.Parameter.Value))
		if err != nil {
			return nil, errors.New("unable to unmarshal value as obj")
		}

		for k, v := range valueMap {
			uKey := *conf.Parameter.Name + ps.config.Delim + k

			// Optionally transform the key.
			if ps.cb != nil {
				uKey = ps.cb(uKey)
			}
			if uKey == "" {
				return nil, errors.New("transformed key is empty")
			}

			mp[uKey] = v
		}
	}

	// Set the response configuration version as the current configuration version.
	// Useful for Watch().
	ps.config.VersionId = strconv.FormatInt(conf.Parameter.Version, 10)

	return maps.Unflatten(mp, ps.config.Delim), nil
}

// ReadBytes returns the raw bytes for parsing.
func (ps *ParameterStore) ReadBytes() ([]byte, error) {
	return nil, errors.New("parameterstore provider does not support this method")
}

// Watch polls AWS AppConfig for configuration updates.
func (ps *ParameterStore) Watch(cb func(event interface{}, err error)) error {
	if ps.config.WatchInterval.Seconds() < 1 {
		ps.config.WatchInterval = 600 * time.Second
	}

	go func() {
	loop:
		for {
			conf, err := ps.client.GetParameter(context.TODO(), &ps.input)
			if err != nil {
				cb(nil, err)
				break loop
			}

			// Check if the the configuration has been updated.
			if strconv.FormatInt(conf.Parameter.Version, 10) == ps.config.VersionId {
				// Configuration is not updated and we have the latest version.
				// Sleep for WatchInterval and retry watcher.
				time.Sleep(ps.config.WatchInterval)
				continue
			}

			// Trigger event.
			cb(nil, nil)
		}
	}()

	return nil
}
