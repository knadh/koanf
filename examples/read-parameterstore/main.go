package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/knadh/koanf/providers/parameterstore/v2"
	"github.com/knadh/koanf/v2"
)

var k = koanf.New(".")

func main() {
	// The configuration values are read from the environment variables.
	c, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	client := ssm.NewFromConfig(c)

	for k, v := range map[string]string{
		"parent1":                    "alice",
		"parent2.child1":             "bob",
		"parent2.child2.grandchild1": "carol",
	} {
		if _, err := client.PutParameter(context.TODO(), &ssm.PutParameterInput{
			Name:      aws.String(k),
			Value:     aws.String(v),
			Type:      types.ParameterTypeSecureString,
			Overwrite: aws.Bool(true),
		}); err != nil {
			log.Fatal(err)
		}
	}

	// Get a parameter.
	if err := k.Load(parameterstore.ProviderWithClient(parameterstore.Config[ssm.GetParameterInput]{
		Delim: ".",
		Input: ssm.GetParameterInput{Name: aws.String("parent1"), WithDecryption: aws.Bool(true)},
	}, client), nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	fmt.Println(k.Sprint())

	// Get parameters.
	if err := k.Load(parameterstore.ProviderWithClient(parameterstore.Config[ssm.GetParametersInput]{
		Delim: ".",
		Input: ssm.GetParametersInput{Names: []string{"parent1", "parent2.child1"}, WithDecryption: aws.Bool(true)},
	}, client), nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	fmt.Println(k.Sprint())

	// Get parameters by path.
	if err := k.Load(parameterstore.ProviderWithClient(parameterstore.Config[ssm.GetParametersByPathInput]{
		Delim: ".",
		Input: ssm.GetParametersByPathInput{Path: aws.String("/"), WithDecryption: aws.Bool(true)},
	}, client), nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	fmt.Println(k.Sprint())
}
