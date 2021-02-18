// Package awssm implements a koanf.Provider that reads raw bytes
// from AWS Secrets Manager to be used with a koanf.Parser to parse
// into conf maps.
package awssm

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Config for the provider.
type Config struct {
	// (optional) AWS Access Key
	AccessKey string

	// (optional) AWS Secret Key
	SecretKey string

	// (optional) AWS SessionToken
	SessionToken string

	// (optional) AWS region
	Region string

	// Secret Name
	SecretName string
}

// AwsSecretsManager implements a AWS Secrets Manager provider.
type AwsSecretsManager struct {
	client *secretsmanager.Client
	cfg    Config
}

// Provider returns a provider that takes a simples3 config.
func Provider(cfg Config) (*AwsSecretsManager, error) {
	var awsCfg aws.Config
	var err error
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		awsCfg, err = config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken)))
	} else {
		awsCfg, err = config.LoadDefaultConfig(context.TODO())
	}
	if err != nil {
		return nil, err
	}
	awsCfg.Region = cfg.Region

	client := secretsmanager.NewFromConfig(awsCfg)
	return &AwsSecretsManager{client: client, cfg: cfg}, nil
}

// ReadBytes reads the contents of a file on s3 and returns the bytes of the json value.
func (r *AwsSecretsManager) ReadBytes() ([]byte, error) {
	s, err := r.client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(r.cfg.SecretName),
	})
	if err != nil {
		return nil, err
	}
	return []byte(*s.SecretString), nil
}

// Read returns the raw bytes for parsing.
func (r *AwsSecretsManager) Read() (map[string]interface{}, error) {
	return nil, errors.New("AwsSecretsManager provider does not support this method")
}

// Watch is not supported.
func (r *AwsSecretsManager) Watch(cb func(event interface{}, err error)) error {
	return errors.New("AwsSecretsManager provider does not support this method")
}
