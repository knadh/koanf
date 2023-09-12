package parameterstore

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
)

func TestParameterStore(t *testing.T) {
	c, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithAPIOptions([]func(*middleware.Stack) error{
			// Mock the SDK response using the middleware.
			func(stack *middleware.Stack) error {
				type key struct{}
				err := stack.Initialize.Add(
					middleware.InitializeMiddlewareFunc(
						"MockInitialize",
						func(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (out middleware.InitializeOutput, metadata middleware.Metadata, err error) {
							switch v := in.Parameters.(type) {
							case *ssm.GetParametersByPathInput:
								ctx = middleware.WithStackValue(ctx, key{}, v.NextToken)
							}
							return next.HandleInitialize(ctx, in)
						},
					), middleware.Before,
				)
				if err != nil {
					return err
				}
				return stack.Finalize.Add(
					middleware.FinalizeMiddlewareFunc(
						"MockFinalize",
						func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
							switch awsmiddleware.GetOperationName(ctx) {
							case "GetParameter":
								return middleware.FinalizeOutput{
									Result: &ssm.GetParameterOutput{
										Parameter: &types.Parameter{
											Name:  aws.String("prefix.parent1"),
											Value: aws.String("alice"),
										},
									},
								}, middleware.Metadata{}, nil
							case "GetParameters":
								return middleware.FinalizeOutput{
									Result: &ssm.GetParametersOutput{
										Parameters: []types.Parameter{
											{
												Name:  aws.String("prefix.parent1"),
												Value: aws.String("alice"),
											},
											{
												Name:  aws.String("prefix.parent2.child1"),
												Value: aws.String("bob"),
											},
										},
									},
								}, middleware.Metadata{}, nil
							case "GetParametersByPath":
								var output ssm.GetParametersByPathOutput
								if middleware.GetStackValue(ctx, key{}) == (*string)(nil) {
									output = ssm.GetParametersByPathOutput{
										NextToken: aws.String("nextToken"),
										Parameters: []types.Parameter{
											{
												Name:  aws.String("prefix.parent1"),
												Value: aws.String("alice"),
											},
											{
												Name:  aws.String("prefix.parent2.child1"),
												Value: aws.String("bob"),
											},
										},
									}
								} else {
									output = ssm.GetParametersByPathOutput{
										Parameters: []types.Parameter{
											{
												Name:  aws.String("prefix.parent2.child2.grandchild1"),
												Value: aws.String("carol"),
											},
										},
									}
								}
								return middleware.FinalizeOutput{Result: &output}, middleware.Metadata{}, nil
							default:
								return middleware.FinalizeOutput{}, middleware.Metadata{}, nil
							}
						},
					),
					middleware.Before,
				)
			},
		}),
	)
	assert.NoError(t, err)
	client := ssm.NewFromConfig(c)

	tests := map[string]struct {
		provider koanf.Provider
		want     map[string]interface{}
	}{
		"get a parameter": {
			provider: ProviderWithClient(Config[ssm.GetParameterInput]{
				Delim:    ".",
				Input:    ssm.GetParameterInput{Name: aws.String("parent1")},
				Callback: func(key, value string) (string, interface{}) { return strings.TrimPrefix(key, "prefix."), value },
			}, client),
			want: map[string]interface{}{
				"parent1": "alice",
			},
		},
		"get parameters": {
			provider: ProviderWithClient(Config[ssm.GetParametersInput]{
				Delim:    ".",
				Input:    ssm.GetParametersInput{Names: []string{"parent1", "parent2.child1"}},
				Callback: func(key, value string) (string, interface{}) { return strings.TrimPrefix(key, "prefix."), value },
			}, client),
			want: map[string]interface{}{
				"parent1": "alice",
				"parent2": map[string]interface{}{
					"child1": "bob",
				},
			},
		},
		"get parameters by path": {
			provider: ProviderWithClient(Config[ssm.GetParametersByPathInput]{
				Delim:    ".",
				Input:    ssm.GetParametersByPathInput{Path: aws.String("/")},
				Callback: func(key, value string) (string, interface{}) { return strings.TrimPrefix(key, "prefix."), value },
			}, client),
			want: map[string]interface{}{
				"parent1": "alice",
				"parent2": map[string]interface{}{
					"child1": "bob",
					"child2": map[string]interface{}{
						"grandchild1": "carol",
					},
				},
			},
		},
		"get a parameter but it is ignored": {
			provider: ProviderWithClient(Config[ssm.GetParameterInput]{
				Delim: ".",
				Input: ssm.GetParameterInput{Name: aws.String("parent1")},
				Callback: func(key, value string) (string, interface{}) {
					return strings.TrimPrefix(strings.TrimPrefix(key, "prefix."), "parent1"), value
				},
			}, client),
			want: map[string]interface{}{
				// Ignored.
				// "parent1": "alice",
			},
		},
		"get parameters but one is ignored": {
			provider: ProviderWithClient(Config[ssm.GetParametersInput]{
				Delim: ".",
				Input: ssm.GetParametersInput{Names: []string{"parent1", "parent2.child1"}},
				Callback: func(key, value string) (string, interface{}) {
					return strings.TrimPrefix(strings.TrimPrefix(key, "prefix."), "parent2.child1"), value
				},
			}, client),
			want: map[string]interface{}{
				"parent1": "alice",
				// Ignored.
				// "parent2": map[string]interface{}{
				// 	"child1": "bob",
				// },
			},
		},
		"get parameters by path but ont is ignored": {
			provider: ProviderWithClient(Config[ssm.GetParametersByPathInput]{
				Delim: ".",
				Input: ssm.GetParametersByPathInput{Path: aws.String("/")},
				Callback: func(key, value string) (string, interface{}) {
					return strings.TrimPrefix(strings.TrimPrefix(key, "prefix."), "parent2.child2.grandchild1"), value
				},
			}, client),
			want: map[string]interface{}{
				"parent1": "alice",
				"parent2": map[string]interface{}{
					"child1": "bob",
					// Ignored.
					// "child2": map[string]interface{}{
					// 	"grandchild1": "carol",
					// },
				},
			},
		},
	}
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			got, err := test.provider.Read()
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
