package awssm

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/stretchr/testify/suite"
)

type AwsSecretsManagerTestSuite struct {
	suite.Suite
	client     *secretsmanager.Client
	secretName string
}

func (suite *AwsSecretsManagerTestSuite) SetupTest() {
	awsCfg, err := config.LoadDefaultConfig(context.TODO())
	awsCfg.Region = "us-west-2"
	suite.NoError(err)
	suite.client = secretsmanager.NewFromConfig(awsCfg)

	suite.secretName = "nonprod/auto-test-" + time.Now().Format("20060102150405")
	_, err = suite.client.CreateSecret(context.TODO(), &secretsmanager.CreateSecretInput{
		Name:         aws.String(suite.secretName),
		SecretString: aws.String(`{"foo":"ping", "bar":"pong"}`),
	},
	)
	suite.NoError(err)
}

func (suite *AwsSecretsManagerTestSuite) TearDownTest() {
	suite.client.DeleteSecret(context.TODO(), &secretsmanager.DeleteSecretInput{
		SecretId: aws.String(suite.secretName),
	})
}

func (suite *AwsSecretsManagerTestSuite) TestExample() {
	// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
	var k = koanf.New(".")

	s, err := Provider(Config{
		SecretName: suite.secretName,
		Region:     "us-west-2",
	})
	suite.NoError(err)

	// Load JSON config.
	err = k.Load(s, json.Parser())
	suite.NoError(err)
	suite.Equal("ping", k.String("foo"))
	suite.Equal("pong", k.String("bar"))
}

func TestAwsSecretsManagerTestSuite(t *testing.T) {
	t.SkipNow()

	suite.Run(t, new(AwsSecretsManagerTestSuite))
}
