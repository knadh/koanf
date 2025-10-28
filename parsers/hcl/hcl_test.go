package hcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHCL_Unmarshal(t *testing.T) {

	hclParserWithFlatten := Parser(true)
	hclParserWithoutFlatten := Parser(false)

	testCases := []struct {
		name     string
		input    []byte
		output   map[string]any
		isErr    bool
		function HCL
	}{
		{
			name:     "Empty HCL - With flatten",
			input:    []byte(`{}`),
			function: *hclParserWithFlatten,
			output:   map[string]any{},
		},
		{
			name:     "Empty HCL - Without flatten",
			input:    []byte(`{}`),
			function: *hclParserWithoutFlatten,
			output:   map[string]any{},
		},
		{
			name: "Valid HCL - With flatten",
			input: []byte(`resource "aws_instance" "example" {
				count = 2 # meta-argument first
				ami           = "abc123"
				instance_type = "t2.micro"
				lifecycle { # meta-argument block last
				  create_before_destroy = true
				}
			  }`),
			function: *hclParserWithFlatten,
			output: map[string]any{
				"resource": map[string]any{
					"aws_instance": map[string]any{
						"example": map[string]any{
							"ami":           "abc123",
							"count":         2,
							"instance_type": "t2.micro",
							"lifecycle": map[string]any{
								"create_before_destroy": true,
							},
						},
					},
				},
			},
		},
		{
			name: "Valid HCL - Without flatten",
			input: []byte(`resource "aws_instance" "example" {
				count = 2 # meta-argument first
				ami           = "abc123"
				instance_type = "t2.micro"
				lifecycle { # meta-argument block last
				  create_before_destroy = true
				}
			  }`),
			function: *hclParserWithoutFlatten,
			output: map[string]any{
				"resource": []map[string]any{{
					"aws_instance": []map[string]any{{
						"example": []map[string]any{{
							"ami":           "abc123",
							"count":         2,
							"instance_type": "t2.micro",
							"lifecycle": []map[string]any{{
								"create_before_destroy": true},
							},
						}},
					}},
				}},
			},
		},
		{
			name: "Invalid HCL - With missing parenthesis",
			input: []byte(`resource "aws_instance" "example" {
				ami = "abc123"
				`),
			function: *hclParserWithFlatten,
			isErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.function.Unmarshal(tc.input)
			if tc.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.output, out)
			}
		})
	}
}

func TestHCL_Marshal(t *testing.T) {

	hclParserWithFlatten := Parser(true)

	testCases := []struct {
		name     string
		input    map[string]any
		isErr    bool
		function HCL
	}{
		{
			name:     "Empty HCL",
			input:    map[string]any{},
			isErr:    true,
			function: *hclParserWithFlatten,
		},
		{
			name: "Complex HCL",
			input: map[string]any{
				"resource": []map[string]any{{
					"aws_instance": []map[string]any{{
						"example": []map[string]any{{
							"ami":           "abc123",
							"count":         2,
							"instance_type": "t2.micro",
							"lifecycle": []map[string]any{{
								"create_before_destroy": true},
							},
						}},
					}},
				}},
			},
			isErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.function.Marshal(tc.input)
			if tc.isErr {
				assert.EqualError(t, err, "HCL marshalling is not supported")
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
