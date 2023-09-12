// Package parameterstore implements a koanf.Provider for AWS Systems Manager Parameter Store.
package parameterstore

import (
	"context"
	"errors"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/knadh/koanf/maps"
)

// Input is a constraint that permits any input type
// to get paramers from AWS Systems Manager Parameter Store.
type Input interface {
	ssm.GetParameterInput | ssm.GetParametersInput | ssm.GetParametersByPathInput
}

// Config represents a ParameterStore provider configuration.
type Config[T Input] struct {
	// Delim is the delimiter to use
	// when specifying config key paths, for instance a . for `parent.child.key`
	// or a / for `parent/child/key`.
	Delim string

	// Input is the input to get parameters.
	Input T

	// OptFns is the additional functional options to get parameters.
	OptFns []func(*ssm.Options)

	// Callback is an optional callback that takes a (key, value)
	// with the variable name and value and allows you to modify both.
	// If the callback returns an empty key, the variable will be ignored.
	Callback func(key, value string) (string, interface{})
}

// ParameterStore implements an AWS Systems Manager Parameter Store provider.
type ParameterStore[T Input] struct {
	client *ssm.Client
	config Config[T]
}

// Provider returns a ParameterStore provider.
// The AWS Systems Manager Client is configured via environment variables.
// The configuration values are read from the environment variables.
//   - AWS_REGION
//   - AWS_ACCESS_KEY_ID
//   - AWS_SECRET_ACCESS_KEY
//   - AWS_SESSION_TOKEN
func Provider[T Input](config Config[T]) *ParameterStore[T] {
	c, err := awsconfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil
	}
	return ProviderWithClient[T](config, ssm.NewFromConfig(c))
}

// ProviderWithClient returns a ParameterStore provider
// using an existing AWS Systems Manager client.
func ProviderWithClient[T Input](config Config[T], client *ssm.Client) *ParameterStore[T] {
	return &ParameterStore[T]{
		client: client,
		config: config,
	}
}

// ReadBytes is not supported by the ParameterStore provider.
func (ps *ParameterStore[T]) ReadBytes() ([]byte, error) {
	return nil, errors.New("parameterstore provider does not support this method")
}

// Read returns a nested config map.
func (ps *ParameterStore[T]) Read() (map[string]interface{}, error) {
	var (
		mp = make(map[string]interface{})
	)
	switch input := interface{}(ps.config.Input).(type) {
	case ssm.GetParameterInput:
		output, err := ps.client.GetParameter(context.TODO(), &input, ps.config.OptFns...)
		if err != nil {
			return nil, err
		}
		// If there's a transformation callback, run it.
		if ps.config.Callback != nil {
			name, value := ps.config.Callback(*output.Parameter.Name, *output.Parameter.Value)
			// If the callback blanked the key, it should be omitted.
			if name == "" {
				break
			}
			mp[name] = value
		} else {
			mp[*output.Parameter.Name] = *output.Parameter.Value
		}
	case ssm.GetParametersInput:
		output, err := ps.client.GetParameters(context.TODO(), &input, ps.config.OptFns...)
		if err != nil {
			return nil, err
		}
		for _, p := range output.Parameters {
			// If there's a transformation callback, run it.
			if ps.config.Callback != nil {
				name, value := ps.config.Callback(*p.Name, *p.Value)
				// If the callback blanked the key, it should be omitted.
				if name == "" {
					break
				}
				mp[name] = value
			} else {
				mp[*p.Name] = *p.Value
			}
		}
	case ssm.GetParametersByPathInput:
		var nextToken *string
		for {
			input.NextToken = nextToken
			output, err := ps.client.GetParametersByPath(context.TODO(), &input, ps.config.OptFns...)
			if err != nil {
				return nil, err
			}
			for _, p := range output.Parameters {
				// If there's a transformation callback, run it.
				if ps.config.Callback != nil {
					name, value := ps.config.Callback(*p.Name, *p.Value)
					// If the callback blanked the key, it should be omitted.
					if name == "" {
						break
					}
					mp[name] = value
				} else {
					mp[*p.Name] = *p.Value
				}
			}
			if output.NextToken == nil {
				break
			}
			nextToken = output.NextToken
		}
	}
	// Unflatten only when a delimiter is specified.
	if ps.config.Delim != "" {
		mp = maps.Unflatten(mp, ps.config.Delim)
	}
	return mp, nil
}
